package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	//"log"
	//"encoding/json"
	//"regexp"
	//"math"
	//"strings"
	//	"os"
	"time"
	//"strings"
	//"bytes"
	//"github.com/DayLightProject/go-daylight/packages/consts"
	//	"math"
	"bytes"
	"database/sql"
)

/* Если майнера забанил админ после того, как к нему пришел запрос cash_request_out,
 * то он всё равно должен отдать свои обещанные суммы, которые получат статус repaid.
 */
func (p *Parser) CashRequestInInit() error {

	fields := []map[string]string{{"cash_request_id": "int64"}, {"code": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

/* не забываем, что cash_request_OUT_front проверяет формат amount,
 * можно ли делать запрос указанному юзеру, есть ли у юзера
 * обещанные суммы на сумму amount, есть ли нужное кол-во DC у отправителя,
 * является ли отправитель майнером
 *
 * */
func (p *Parser) CashRequestInFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// code может быть чем угодно, т.к. отправитель шлет в сеть лишь хэш
	// нигде, кроме cash_request_in_front, code не используется
	// if ( !check_input_data ($this->tx_data['code'], 'cash_code') )
	//	return 'cash_request_in_front code';

	verifyData := map[string]string{"cash_request_id": "bigint"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	var to_user_id, cTime int64
	var status string
	var hash_code []byte
	err = p.QueryRow(p.FormatQuery("SELECT to_user_id, status, hash_code, time FROM cash_requests WHERE id  =  ?"), p.TxMaps.Int64["cash_request_id"]).Scan(&to_user_id, &status, &hash_code, &cTime)
	if err != nil && err != sql.ErrNoRows {
		return p.ErrInfo(err)
	}

	// ID cash_requests юзер указал сам, значит это может быть случайное число.
	// проверим, является получателем наш юзер
	if to_user_id != p.TxUserID {
		return p.ErrInfo("to_user_id!=user_id")
	}
	// должно быть pending
	if status != "pending" {
		return p.ErrInfo("status!=pending")
	}
	// проверим код
	if !bytes.Equal(utils.DSha256(p.TxMaps.String["code"]), utils.BinToHex(hash_code)) {
		return p.ErrInfo("code!=hash_code")
	}
	var txTime int64
	if p.BlockData != nil { // тр-ия пришла в блоке
		txTime = p.BlockData.Time
	} else {
		txTime = time.Now().Unix() // просто на всякий случай небольшой запас
	}
	// запрос может быть принят, только если он был отправлен не позднее чем через cash_request_time сек назад
	if cTime < txTime-p.Variables.Int64["cash_request_time"] {
		return p.ErrInfo(fmt.Sprintf("%d < %d - %d", cTime, txTime, p.Variables.Int64["cash_request_time"]))
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["cash_request_id"], p.TxMap["code"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	return nil
}

func (p *Parser) CashRequestIn() error {
	var to_user_id, from_user_id, currency_id, cTime int64
	var status string
	var hash_code []byte
	var amount float64
	err := p.QueryRow(p.FormatQuery("SELECT from_user_id, to_user_id, currency_id, status, hash_code, time, amount FROM cash_requests WHERE id  =  ?"), p.TxMaps.Int64["cash_request_id"]).Scan(&from_user_id, &to_user_id, &currency_id, &status, &hash_code, &cTime, &amount)
	if err != nil && err != sql.ErrNoRows {
		return p.ErrInfo(err)
	}
	// возможно нужно обновить таблицу points_status
	err = p.pointsUpdateMain(from_user_id)
	if err != nil {
		return p.ErrInfo(err)
	}
	promisedAmountStatus := "repaid"
	// есть вероятность того, что после попадания в Dc-сеть cash_request_out придет admin_ban_miner, а после попадения в сеть cash_request_in придет admin_unban_miner. В admin_unban_miner смена статуса suspended на repaid у нового promised_amount учтено
	userStatus, err := p.Single("SELECT status FROM miners_data WHERE user_id  =  ?", p.TxUserID).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	var repaidPromisedAmountId int64
	if userStatus == "suspended_miner" {
		promisedAmountStatus = "suspended"
		// нужно понять, какой promised_amount ранее имел статус repaid
		repaidPromisedAmountId, err = p.Single("SELECT id FROM promised_amount WHERE user_id  =  ? AND currency_id  =  ? AND status_backup  =  'repaid' AND del_block_id  =  0 AND del_mining_block_id  =  0", p.TxUserID, currency_id).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
	} else {
		// ну а если майнер не забанен админом, то всё просто
		repaidPromisedAmountId, err = p.Single("SELECT id FROM promised_amount WHERE user_id  =  ? AND currency_id  =  ? AND status  =  'repaid' AND del_block_id  =  0 AND del_mining_block_id  =  0", p.TxUserID, currency_id).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	// если уже есть repaid для данной валюты, то просто приплюсуем к сумме
	if repaidPromisedAmountId > 0 {
		data, err := p.OneRow("SELECT * FROM promised_amount WHERE id  =  ?", repaidPromisedAmountId).String()
		if err != nil {
			return p.ErrInfo(err)
		}
		logId, err := p.ExecSqlGetLastInsertId("INSERT INTO log_promised_amount ( amount, tdc_amount, tdc_amount_update, cash_request_in_block_id, block_id, prev_log_id ) VALUES ( ?, ?, ?, ?, ?, ? )", "log_id", data["amount"], data["tdc_amount"], data["tdc_amount_update"], data["cash_request_in_block_id"], p.BlockData.BlockId, data["log_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		// tdc_amount не пересчитываются, т.к. пока есть cash_requests с pending, они не растут
		err = p.ExecSql("UPDATE promised_amount SET amount = amount + ?, tdc_amount = ?, tdc_amount_update = ?, cash_request_in_block_id = ?, log_id = ? WHERE id = ?", amount, (utils.StrToFloat64(data["tdc_amount"]) + amount), p.BlockData.Time, p.BlockData.BlockId, logId, repaidPromisedAmountId)
		if err != nil {
			return p.ErrInfo(err)
		}
	} else {
		err = p.ExecSql("INSERT INTO promised_amount ( user_id, amount, currency_id, start_time, status, tdc_amount, tdc_amount_update, cash_request_in_block_id ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ? )", p.TxUserID, amount, currency_id, p.BlockData.Time, promisedAmountStatus, amount, p.BlockData.Time, p.BlockData.BlockId)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	// теперь нужно вычесть зачисленную сумму на repaid из mining
	data, err := p.OneRow("SELECT * FROM promised_amount WHERE user_id  =  ? AND currency_id  =  ? AND status  =  'mining' AND del_block_id  =  0 AND del_mining_block_id  =  0", p.TxUserID, currency_id).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	logId, err := p.ExecSqlGetLastInsertId("INSERT INTO log_promised_amount ( amount, tdc_amount, tdc_amount_update, cash_request_in_block_id, block_id, prev_log_id ) VALUES ( ?, ?, ?, ?, ?, ? )", "log_id", data["amount"], data["tdc_amount"], data["tdc_amount_update"], data["cash_request_in_block_id"], p.BlockData.BlockId, data["log_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// вычитаем из mining то, что начислили выше на repaid
	// tdc_amount не пересчитываются, т.к. пока есть cash_requests с pending, они не растут
	err = p.ExecSql("UPDATE promised_amount SET amount = amount - ?, tdc_amount = ?, tdc_amount_update = ?, cash_request_in_block_id = ?, log_id = ? WHERE id = ?", amount, data["tdc_amount"], p.BlockData.Time, p.BlockData.BlockId, logId, data["id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// обновим сумму на кошельке отправителя, вычтя amount и залогировав предыдущее значение
	err = p.updateSenderWallet(from_user_id, currency_id, amount, 0, "cash_request", p.TxMaps.Int64["cash_request_id"], p.TxUserID, "cash_request", "decrypted")
	if err != nil {
		return p.ErrInfo(err)
	}

	// Отмечаем, что данный cash_requests погашен.
	err = p.ExecSql("UPDATE cash_requests SET status = 'approved' WHERE id = ?", p.TxMaps.Int64["cash_request_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// возможно, больше нет mining ни по одной валюте (кроме WOC) у данного юзера
	forRepaidCurrencyIds, err := p.GetList("SELECT currency_id FROM promised_amount WHERE status  =  'mining' AND user_id  =  ? AND amount > 0 AND currency_id > 1 AND del_block_id  =  0 AND del_mining_block_id  =  0", p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	var forRepaidCurrencyIdsNew []int64
	for _, currencyId := range forRepaidCurrencyIds {
		// либо сумма погашенных стала >= максимальной обещанной, т.к. в этом случае прислать этому юзеру cash_request_out будет невозможно
		maxPromisedAmount, err := p.GetMaxPromisedAmount(currencyId)
		if err != nil {
			return p.ErrInfo(err)
		}
		repaidAmount, err := p.GetRepaidAmount(currencyId, p.TxUserID)
		if err != nil {
			return p.ErrInfo(err)
		}
		if repaidAmount < maxPromisedAmount {
			forRepaidCurrencyIdsNew = append(forRepaidCurrencyIdsNew, currencyId)
		}
	}
	if len(forRepaidCurrencyIdsNew) == 0 {
		// просроченным cash_requests ставим for_repaid_del_block_id, чтобы было ясно, что юзер не имеет долгов, и его TDC должны расти
		err = p.ExecSql("UPDATE cash_requests SET for_repaid_del_block_id = ? WHERE to_user_id = ? AND time < ? AND for_repaid_del_block_id = 0", p.BlockData.BlockId, p.TxUserID, (p.BlockData.Time - p.Variables.Int64["cash_request_time"]))
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	existsRequests := p.CheckCashRequests(p.TxUserID)
	// возможно, что данный cash_requests с approved был единственный, и последующий вызов метода mining начислит новые TDC в соответствии с имеющимся % роста,. значит необходимо обновить tdc_amount и tdc_amount_update
	if len(forRepaidCurrencyIdsNew) == 0 || existsRequests == nil { // у юзера нет долгов, нужно ставить ему cash_request_out_time=0
		err = p.updPromisedAmounts(p.TxUserID, false, true, 0)
		if err != nil {
			return p.ErrInfo(err)
		}
	} else {
		// для того, чтобы было проще делать rollback пишем время cash_request_out_time, хотя по сути cash_request_out_time будет таким же каким и был
		cashRequestOutTime, err := p.Single("SELECT cash_request_out_time FROM promised_amount WHERE user_id  =  ? AND cash_request_out_time > 0", p.TxUserID).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.updPromisedAmounts(p.TxUserID, false, true, cashRequestOutTime)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	cashRequestsDataFromUserId, err := p.Single("SELECT from_user_id FROM cash_requests WHERE id  =  ?", p.TxMaps.Int64["cash_request_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	// проверим, не наш ли это user_id
	_, myBlockId, myPrefix, myUserIds, err := p.GetMyUserId(p.TxMaps.Int64["to_user_id"])
	if err != nil {
		return err
	}
	if (utils.InSliceInt64(p.TxUserID, myUserIds) || utils.InSliceInt64(cashRequestsDataFromUserId, myUserIds)) && myBlockId <= p.BlockData.BlockId {
		collective, err := p.GetCommunityUsers()
		if err != nil {
			return err
		}
		if len(collective) > 0 && utils.InSliceInt64(cashRequestsDataFromUserId, myUserIds) { // наш юзер - это отправитель _out
			myPrefix = utils.Int64ToStr(cashRequestsDataFromUserId) + "_"
		} else if len(collective) > 0 && utils.InSliceInt64(p.TxUserID, myUserIds) { // наш юзер - это отправитель _in
			myPrefix = utils.Int64ToStr(p.TxUserID) + "_"
		} else {
			myPrefix = ""
		}
		// обновим таблу, отметив, что мы отдали деньги
		err = p.ExecSql("UPDATE "+myPrefix+"my_cash_requests SET status = 'approved' WHERE cash_request_id = ?", p.TxMaps.Int64["cash_request_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	return nil
}

func (p *Parser) CashRequestInRollback() error {
	err := p.updPromisedAmountsRollback(p.TxUserID, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql("UPDATE cash_requests SET for_repaid_del_block_id = 0 WHERE to_user_id = ? AND for_repaid_del_block_id = ?", p.TxUserID, p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	var to_user_id, from_user_id, cTime int64
	var status, currency_id string
	var hash_code []byte
	var amount float64
	err = p.QueryRow(p.FormatQuery("SELECT from_user_id, to_user_id, currency_id, status, hash_code, time, amount FROM cash_requests WHERE id  =  ?"), p.TxMaps.Int64["cash_request_id"]).Scan(&from_user_id, &to_user_id, &currency_id, &status, &hash_code, &cTime, &amount)
	if err != nil && err != sql.ErrNoRows {
		return p.ErrInfo(err)
	}

	err = p.pointsUpdateRollbackMain(from_user_id)
	if err != nil {
		return p.ErrInfo(err)
	}

	// откатим cash_requests
	err = p.ExecSql("UPDATE cash_requests SET status = 'pending' WHERE id = ?", p.TxMaps.Int64["cash_request_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// откатим DC, списанные с кошелька отправителя DC
	err = p.generalRollback("wallets", from_user_id, "AND currency_id = "+currency_id, false)
	if err != nil {
		return p.ErrInfo(err)
	}

	// откатываем обещанные суммы, у которых было затронуто amount
	rows, err := p.Query(p.FormatQuery("SELECT id, log_id FROM promised_amount WHERE user_id = ? AND currency_id = ? AND cash_request_in_block_id = ? AND del_block_id = 0 AND del_mining_block_id = 0 ORDER BY log_id DESC"), p.TxUserID, currency_id, p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id, log_id int64
		err = rows.Scan(&id, &log_id)
		if err != nil {
			return p.ErrInfo(err)
		}
		if log_id > 0 {
			logData, err := p.OneRow("SELECT amount, tdc_amount, tdc_amount_update, cash_request_in_block_id, prev_log_id FROM log_promised_amount WHERE log_id  =  ?", log_id).String()
			if err != nil {
				return p.ErrInfo(err)
			}
			err = p.ExecSql("UPDATE promised_amount SET amount = ?, tdc_amount = ?, tdc_amount_update = ?,  cash_request_in_block_id = ?, log_id = ? WHERE id = ?", logData["amount"], logData["tdc_amount"], logData["tdc_amount_update"], logData["cash_request_in_block_id"], logData["prev_log_id"], id)
			if err != nil {
				return p.ErrInfo(err)
			}
			err = p.ExecSql("DELETE FROM log_promised_amount WHERE log_id = ?", log_id)
			if err != nil {
				return p.ErrInfo(err)
			}
			err = p.rollbackAI("log_promised_amount", 1)
			if err != nil {
				return p.ErrInfo(err)
			}

		} else {
			err = p.ExecSql("DELETE FROM promised_amount WHERE id = ?", id)
			if err != nil {
				return p.ErrInfo(err)
			}
			err = p.rollbackAI("promised_amount", 1)
			if err != nil {
				return p.ErrInfo(err)
			}
		}
	}

	cashRequestsFromUserId, err := p.Single("SELECT from_user_id FROM cash_requests WHERE id  =  ?", p.TxMaps.Int64["cash_request_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	// проверим, не наш ли это user_id
	_, _, myPrefix, myUserIds, err := p.GetMyUserId(p.TxUserID)
	if err != nil {
		return err
	}
	if utils.InSliceInt64(p.TxUserID, myUserIds) || utils.InSliceInt64(cashRequestsFromUserId, myUserIds) {
		collective, err := p.GetCommunityUsers()
		if err != nil {
			return err
		}
		if len(collective) > 0 && utils.InSliceInt64(cashRequestsFromUserId, myUserIds) { // наш юзер - это отправитель _out
			myPrefix = utils.Int64ToStr(cashRequestsFromUserId) + "_"
		} else if len(collective) > 0 && utils.InSliceInt64(p.TxUserID, myUserIds) { // наш юзер - это отправитель _in
			myPrefix = utils.Int64ToStr(p.TxUserID) + "_"
		} else {
			myPrefix = ""
		}
		// обновим таблу, отметив, что мы отдали деньги
		err = p.ExecSql("UPDATE "+myPrefix+"my_cash_requests SET status = 'approved' WHERE cash_request_id = ?", p.TxMaps.Int64["cash_request_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		if utils.InSliceInt64(cashRequestsFromUserId, myUserIds) {
			err = p.ExecSql("DELETE FROM "+myPrefix+"my_dc_transactions WHERE status = 'approved' AND type = 'cash_request' AND amount = ? AND block_id = ? AND currency_id = ?", amount, p.BlockData.BlockId, currency_id)
			if err != nil {
				return p.ErrInfo(err)
			}
		}
	}
	err = p.mydctxRollback()
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) CashRequestInRollbackFront() error {

	return nil

}
