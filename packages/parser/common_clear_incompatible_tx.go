// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package parser

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) ClearIncompatibleTx(binaryTx []byte, myTx bool) (string, string, int64, int64, int64, int64, int64) {

	var fatalError, waitError string
	var thirdVar int64

	// получим тип тр-ии и юзера
	txType, walletId, citizenId := utils.GetTxTypeAndUserId(binaryTx)

	if !utils.CheckInputData(txType, "int") {
		fatalError = "error type"
	}
	if !utils.CheckInputData(walletId, "int64") {
		fatalError = "error walletId"
	}
	if !utils.CheckInputData(citizenId, "int") {
		fatalError = "error citizenId"
	}

	var forSelfUse int64

	/*
		// две тр-ии одного типа от одного юзера не должны попасть в один блок
		// исключение - перевод DC между юзерами
		if len(fatalError) == 0 {
			p.ClearIncompatibleTxSql(txType, walletId, citizenId, &waitError)

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
						`, citizenId).Int64()
			if err != nil {
				fatalError = fmt.Sprintf("%s", err)
			}
			if num > 0 {
				waitError = "only 1 tx"
			}
		}*/
	log.Debug("fatalError: %v, waitError: %v, forSelfUse: %v, txType: %v, walletId: %v, citizenId: %v, thirdVar: %v", fatalError, waitError, forSelfUse, txType, walletId, citizenId, thirdVar)
	return fatalError, waitError, forSelfUse, txType, walletId, citizenId, thirdVar
}
