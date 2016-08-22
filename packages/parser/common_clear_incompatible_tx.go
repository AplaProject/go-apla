package parser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) ClearIncompatibleTx(binaryTx []byte, myTx bool) (string, string, int64, int64, int64, int64, int64) {

	var fatalError, waitError string
	var toUserId int64

	// получим тип тр-ии и юзера
	txType, walletId, citizenId, thirdVar := utils.GetTxTypeAndUserId(binaryTx)

	if !utils.CheckInputData(txType, "int") {
		fatalError = "error type"
	}
	if !utils.CheckInputData(walletId, "int") {
		fatalError = "error walletId"
	}
	if !utils.CheckInputData(citizenId, "int") {
		fatalError = "error citizenId"
	}
	if !utils.CheckInputData(thirdVar, "int") {
		fatalError = "error thirdVar"
	}


	var forSelfUse int64
	if utils.InSliceInt64(txType, utils.TypesToIds([]string{"NewPct", "NewReduction", "NewMaxPromisedAmounts", "NewMaxOtherCurrencies"})) {
		//  чтобы никому не слать эту тр-ю
		forSelfUse = 1
		// $my_tx == true - это значит функция вызвана из pct_generator reduction_generator
		// если же false, то она была спаршена query_tx или tesblock_generator и имела verified=0
		// а т.к. new_pct/NewReduction актуальны только 1 блок, то нужно её удалять
		if !myTx {
			fatalError = "old new_pct/NewReduction/NewMaxPromisedAmounts/NewMaxOtherCurrencies"
			return fatalError, waitError, forSelfUse, txType, walletId, citizenId, toUserId
		}
	} else {
		forSelfUse = 0
	}

	// две тр-ии одного типа от одного юзера не должны попасть в один блок
	// исключение - перевод DC между юзерами
	if len(fatalError) == 0 {
		p.ClearIncompatibleTxSql(txType, walletId, citizenId, &waitError)


		// нельзя удалять CF-проект и в этом же блоке изменить его описание/профинансировать
		if txType == utils.TypeInt("DelCfProject") {
			p.ClearIncompatibleTxSqlSet([]string{"CfSendDc"}, 0, 0, &waitError, thirdVar)
		}
		if utils.InSliceInt64(txType, utils.TypesToIds([]string{"CfSendDc"})) {
			p.ClearIncompatibleTxSqlSet([]string{"DelCfProject"}, 0, 0, &waitError, thirdVar)
		}

		// потом нужно сделать более тонко. но пока так. Если есть удаление проекта, тогда откатываем все тр-ии del_cf_funding
		if txType == utils.TypeInt("DelCfProject") {
			p.RollbackIncompatibleTx([]string{"DelCfFunding"})
		}

		// Если есть смена коммиссий арбитров, то нельзя делать перевод монет, т.к. там может быть указана комиссия арбитра
		if utils.InSliceInt64(txType, utils.TypesToIds([]string{"SendDc"})) {
			p.RollbackIncompatibleTx([]string{"ChangeArbitratorConditions"})
		}
		if txType == utils.TypeInt("ChangeArbitratorConditions") {
			p.ClearIncompatibleTxSqlSet([]string{"SendDc"}, 0, 0, &waitError, "")
		}


		// на всякий случай не даем попасть в один блок тр-ии отправки в CF-проект монет и другим тр-ям связанным с этим CF-проектом. Т.к. проект может завершиться и 2-я тр-я вызовет ошибку
		if txType == utils.TypeInt("CfSendDc") {
			p.ClearIncompatibleTxSqlSet([]string{"DelCfProject"}, 0, 0, &waitError, thirdVar)
		}
		if utils.InSliceInt64(txType, utils.TypesToIds([]string{"DelCfProject"})) {
			p.ClearIncompatibleTxSqlSet([]string{"CfSendDc"}, 0, 0, &waitError, thirdVar)
		}

		// в один блок должен попасть только один голос за один объект голосования. thirdVar - объект голосования
		if utils.InSliceInt64(txType, utils.TypesToIds([]string{"VotesPromisedAmount", "VotesMiner", "VotesNodeNewMiner", "VotesComplex"})) {
			num, err := p.Single(`
			  			  SELECT count(*)
				            FROM (
					            SELECT citizen_id
					            FROM transactions
					            WHERE  type IN (?, ?, ?, ?) AND
					                          third_var = ? AND
					                          verified=1 AND
					                          used = 0
							)  AS x
							`, utils.TypeInt("VotesPromisedAmount"), utils.TypeInt("VotesMiner"), utils.TypeInt("VotesNodeNewMiner"), utils.TypeInt("VotesComplex"), thirdVar, utils.TypeInt("VotesPromisedAmount"), utils.TypeInt("VotesMiner"), utils.TypeInt("VotesNodeNewMiner"), utils.TypeInt("VotesComplex"), thirdVar).Int64()
			if err != nil {
				fatalError = fmt.Sprintf("%s", err)
			}
			if num > 0 {
				waitError = "only 1 vote"
			}
		}

		// если новая тр-я - это смена праймари ключа, то не должно быть никаких других тр-ий от этого юзера
		if txType == utils.TypeInt("ChangePrimaryKey") {
			num, err := p.Single(`
						  SELECT count(*)
				            FROM (
					            SELECT citizen_id
					            FROM transactions
					            WHERE  user_id = ? AND
					                         verified=1 AND
					                         used = 0
							)  AS x
							`, citizenId, citizenId).Int64()
			if err != nil {
				fatalError = fmt.Sprintf("%s", err)
			}
			if num > 0 {
				waitError = "there are other tr-s"
			}
		}

		// любая тр-я от юзера не должна проходить, если уже есть тр-я со сменой праймари ключа или new_pct или NewReduction
		num, err := p.Single(`
						SELECT count(*)
				          FROM (
					            SELECT citizen_id
					            FROM transactions
					            WHERE  (
						                            (type = ? AND citizen_id = ?)
						                            OR
						                            (type IN (?, ?) )
					                          ) AND
					                         verified=1 AND
					                         used = 0

						)  AS x
						`, utils.TypeInt("ChangePrimaryKey"), citizenId, utils.TypeInt("NewPct"), utils.TypeInt("NewReduction"), utils.TypeInt("ChangePrimaryKey"), citizenId, utils.TypeInt("NewPct"), utils.TypeInt("NewReduction")).Int64()
		if err != nil {
			fatalError = fmt.Sprintf("%s", err)
		}
		if num > 0 {
			waitError = "have ChangePrimaryKey tx"
		}


		// временно запрещаем 2 тр-ии любого типа от одного юзера, а то затрахался уже.
		num, err = p.Single(`
				    SELECT count(*)
				    FROM (
							SELECT citizen_id
							FROM transactions
							WHERE  citizen_id = ? AND
				                      verified=1 AND
				                      used = 0
					)  AS x
					`, citizenId, citizenId).Int64()
		if err != nil {
			fatalError = fmt.Sprintf("%s", err)
		}
		if num > 0 {
			waitError = "only 1 tx"
		}
	}
	log.Debug("fatalError: %v, waitError: %v, forSelfUse: %v, txType: %v, walletId: %v, citizenId: %v, thirdVar: %v", fatalError, waitError, forSelfUse, txType, walletId, citizenId, thirdVar)
	return fatalError, waitError, forSelfUse, txType, walletId, citizenId, thirdVar

}
