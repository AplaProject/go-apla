package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

func (p *Parser) CashRequestOutInit() error {

	fields := []map[string]string{{"to_user_id": "int64"}, {"amount": "money"}, {"comment": "string"}, {"currency_id": "int64"}, {"hash_code": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) CashRequestOutFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"to_user_id": "bigint", "currency_id": "currency_id", "amount": "amount", "comment": "comment", "hash_code": "sha256"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// нельзя слать запрос на woc
	if p.TxMaps.Int64["currency_id"] == 1 {
		return p.ErrInfo("WOC")
	}

	// прошло ли 30 дней с момента регистрации майнера
	minerNewbie := p.checkMinerNewbie()
	if minerNewbie != nil {
		// возможно, что майнер отдал наличные за DC и у него есть общенные суммы с repaid
		// нужно дать возможность вывести ровно столько, сколько он отдал
		repaidAmount, err := p.GetRepaidAmount(p.TxMaps.Int64["currency_id"], p.TxUserID)
		// сколько уже получил наличных
		amountCashRequests, err := p.Single("SELECT sum(amount) FROM cash_requests WHERE status  =  'approved' AND currency_id  =  ? AND from_user_id  =  ?", p.TxMaps.Int64["currency_id"], p.TxUserID).Float64()
		if err != nil {
			return p.ErrInfo(err)
		}
		if amountCashRequests+p.TxMaps.Money["amount"] > repaidAmount {
			return p.ErrInfo(fmt.Sprintf("%f + %f > %f", amountCashRequests, p.TxMaps.Money["amount"], repaidAmount))
		}
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["to_user_id"], p.TxMap["amount"], utils.BinToHex(p.TxMap["comment"]), p.TxMap["currency_id"], p.TxMap["hash_code"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	// проверим, существует ли такая валюта в таблице DC-валют
	if ok, err := p.CheckCurrency(p.TxMaps.Int64["currency_id"]); !ok {
		return p.ErrInfo(err)
	}

	// ===  begin проверка to_user_id

	// является ли данный юзер майнером
	err = p.checkMiner(p.TxMaps.Int64["to_user_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	maxPromisedAmount, err := p.GetMaxPromisedAmount(p.TxMaps.Int64["currency_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	repaidAmount, err := p.GetRepaidAmount(p.TxMaps.Int64["currency_id"], p.TxMaps.Int64["to_user_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	if p.TxMaps.Money["amount"]+repaidAmount > maxPromisedAmount {
		return p.ErrInfo(fmt.Sprintf("%f + %f > %f", p.TxMaps.Money["amount"], repaidAmount, maxPromisedAmount))
	}

	// не даем превысить общий лимит
	promisedAmount, err := p.Single("SELECT amount FROM promised_amount WHERE status  =  'mining' AND currency_id  =  ? AND user_id  =  ? AND del_block_id  =  0 AND del_mining_block_id  =  0", p.TxMaps.Int64["currency_id"], p.TxMaps.Int64["to_user_id"]).Float64()
	if err != nil {
		return p.ErrInfo(err)
	}
	rest := maxPromisedAmount - repaidAmount
	if rest < promisedAmount {
		promisedAmount = rest
	}

	// минимальная сумма. теоретически может делиться на min_promised_amount пока не достигнет 0.01
	if p.TxMaps.Money["amount"] < promisedAmount/float64(p.Variables.Int64["min_promised_amount"]) {
		return p.ErrInfo(fmt.Sprintf("%f < %f / %d", p.TxMaps.Money["amount"], promisedAmount, p.Variables.Int64["min_promised_amount"]))
	}
	if p.TxMaps.Money["amount"] < 0.01 {
		return p.ErrInfo("amount<0.01")
	}

	var txTime int64
	if p.BlockData != nil { // тр-ия пришла в блоке
		txTime = p.BlockData.Time
	} else {
		txTime = time.Now().Unix() - 30 // просто на всякий случай небольшой запас
	}

	// Чтобы не задалбывать получателей запроса на обмен, не даем отправить следующий запрос, пока не пройдет cash_request_time сек с момента предыдущего
	cashRequestPending, err := p.Single("SELECT status FROM cash_requests WHERE to_user_id  =  ? AND del_block_id  =  0 AND for_repaid_del_block_id  =  0 AND time > ? AND status  =  'pending'", p.TxMaps.Int64["to_user_id"], (txTime - p.Variables.Int64["cash_request_time"])).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(cashRequestPending) > 0 {
		return p.ErrInfo("cash_requests status not null")
	}

	// не находится ли юзер в данный момент на каникулах.
	rows, err := p.Query(p.FormatQuery("SELECT start_time, end_time FROM holidays WHERE user_id = ? AND del = 0"), p.TxMaps.Int64["to_user_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var startTime, endTime int64
		err = rows.Scan(&startTime, &endTime)
		if err != nil {
			return p.ErrInfo(err)
		}
		var time1, time2 int64
		if p.BlockData != nil {
			time1 = p.BlockData.Time
			time2 = time1
		} else {
			// тут используем time() с запасом 1800 сек, т.к. в момент, когда тр-ия попадет в блок, каникулы уже могут начаться.
			// т.е. у голой тр-ии проверка идет жестче
			time1 = time.Now().Unix() + 1800
			time2 = time.Now().Unix()
		}
		if startTime <= time1 && endTime >= time2 {
			return p.ErrInfo("error holidays")
		}
	}
	// === end проверка to_user_id

	// ===  begin проверка отправителя
	// является ли данный юзер майнером
	err = p.checkMiner(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	/*
	 * WalletsBuffer тут не используем т.к. попадение 2-х тр-ий данного типа в 1 блок во время генарции блока исключается в clear_incompatible_tx.
	 * там же исключается попадение данного типа с new_forex и пр.
	 * А проверка 2-го списания идет в ParseDataFull
	 * */
	if p.BlockData == nil || p.BlockData.BlockId > 173941 {
		// в блоке 173941 была попытка отправить 2 запроса на 408 и 200 dUSD, в то время как на кошельке было только 449.6
		// в итоге check_sender_money возвращало ошибку
		// есть ли нужная сумма на кошельке
		p.TxMaps.Int64["from_user_id"] = p.TxMaps.Int64["user_id"]
		for i := 0; i < 5; i++ {
			p.TxMaps.Money["arbitrator"+utils.IntToStr(i)+"_commission"] = 0
		}
		p.TxMaps.Money["commission"] = 0
		//func (p *Parser) checkSenderMoney(currencyId, fromUserId int64, amount, commission, arbitrator0_commission, arbitrator1_commission, arbitrator2_commission, arbitrator3_commission, arbitrator4_commission float64) (float64, error) {
		_, err := p.checkSenderMoney(p.TxMaps.Int64["currency_id"], p.TxMaps.Int64["from_user_id"], p.TxMaps.Money["amount"], p.TxMaps.Money["commission"], p.TxMaps.Money["arbitrator0_commission"], p.TxMaps.Money["arbitrator1_commission"], p.TxMaps.Money["arbitrator2_commission"], p.TxMaps.Money["arbitrator3_commission"], p.TxMaps.Money["arbitrator4_commission"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	// У юзера не должно быть cash_requests со статусом pending
	err = p.CheckCashRequests(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.limitRequest(p.Variables.Int64["limit_cash_requests_out"], "cash_requests", p.Variables.Int64["limit_cash_requests_out_period"])
	if err != nil {
		return p.ErrInfo(err)
	}
	// ===  end проверка отправителя

	return nil
}

func (p *Parser) CashRequestOut() error {

	// возможно нужно обновить таблицу points_status
	err := p.pointsUpdateMain(p.TxMaps.Int64["to_user_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// у получателя запроса останавливается майнинг по всем валютам и статусам, т.е. mining/pending. значит необходимо обновить tdc_amount и tdc_amount_update
	// WOC продолжает расти
	// обновление нужно, только если данный cash_request единственный с pending, иначе делать пересчет tdc_amount нельзя, т.к. уже были ранее пересчитаны
	existsRequests := p.CheckCashRequests(p.TxMaps.Int64["to_user_id"])
	if existsRequests == nil {
		log.Debug("updPromisedAmounts=", existsRequests)
		err = p.updPromisedAmounts(p.TxMaps.Int64["to_user_id"], true, true, p.BlockData.Time)
		if err != nil {
			return p.ErrInfo(err)
		}
	} else {
		// записываем cash_request_out_time во всех обещанных суммах, после того, как юзер вызвал актуализацию. акутализацию юзер вызывал т.к. у него есть непогашенные cash_request.
		log.Debug("updPromisedAmountsCashRequestOutTime")
		err = p.updPromisedAmountsCashRequestOutTime(p.TxMaps.Int64["to_user_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	// пишем запрос в БД
	cashRequestId, err := p.ExecSqlGetLastInsertId("INSERT INTO cash_requests ( time, from_user_id, to_user_id, currency_id, amount, hash_code ) VALUES ( ?, ?, ?, ?, ?, [hex] )", "id", p.BlockData.Time, p.TxUserID, p.TxMaps.Int64["to_user_id"], p.TxMaps.Int64["currency_id"], p.TxMaps.Money["amount"], p.TxMaps.String["hash_code"])
	if err != nil {
		return p.ErrInfo(err)
	}
	// а может быть наш юзер - получатель запроса
	myUserId, myBlockId, myPrefix, _, err := p.GetMyUserId(p.TxMaps.Int64["to_user_id"])
	if err != nil {
		return err
	}
	if p.TxMaps.Int64["to_user_id"] == myUserId && myBlockId <= p.BlockData.BlockId {
		// пишем с таблу инфу, что к нам пришел новый запрос
		err = p.ExecSql("INSERT INTO "+myPrefix+"my_cash_requests ( time, to_user_id, currency_id, amount, comment, comment_status, status, hash_code, cash_request_id ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ? )", p.BlockData.Time, p.TxMaps.Int64["to_user_id"], p.TxMaps.Int64["currency_id"], p.TxMaps.Money["amount"], utils.BinToHex(p.TxMap["comment"]), "encrypted", "pending", p.TxMaps.String["hash_code"], cashRequestId)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	// или отправитель запроса - наш юзер
	myUserId, myBlockId, myPrefix, _, err = p.GetMyUserId(p.TxUserID)
	if err != nil {
		return err
	}
	if p.TxUserID == myUserId && myBlockId <= p.BlockData.BlockId {
		myId, err := p.Single("SELECT id FROM "+myPrefix+"my_cash_requests WHERE to_user_id  =  ? AND status  =  'my_pending' ORDER BY id DESC", p.TxMaps.Int64["to_user_id"]).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		if myId > 0 {
			// обновим статус в нашей локальной табле.
			// у юзера может быть только 1 запрос к 1 юзеру со статусом my_pending
			err = p.ExecSql("UPDATE "+myPrefix+"my_cash_requests SET status = 'pending', time = ?, cash_request_id = ? WHERE id = ?", p.BlockData.Time, cashRequestId, myId)
			if err != nil {
				return p.ErrInfo(err)
			}
		} else {
			err = p.ExecSql("INSERT INTO "+myPrefix+"my_cash_requests ( to_user_id, currency_id, amount, comment, hash_code, status, cash_request_id ) VALUES ( ?, ?, ?, ?, ?, ?, ? )", p.TxMaps.Int64["to_user_id"], p.TxMaps.Int64["currency_id"], p.TxMaps.Money["amount"], "", p.TxMaps.String["hash_code"], "pending", cashRequestId)
			if err != nil {
				return p.ErrInfo(err)
			}
		}

		myId, err = p.Single("SELECT id FROM "+myPrefix+"my_dc_transactions WHERE status  =  'pending' AND type  =  'cash_request' AND to_user_id  =  ? AND amount  =  ? AND currency_id  =  ?", p.TxMaps.Int64["to_user_id"], p.TxMaps.Money["amount"], p.TxMaps.Int64["currency_id"]).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		if myId > 0 {
			// чтобы при вызове update_sender_wallet из cash_request_in можно было обновить my_dc_transactions, т.к. там в WHERE есть type_id
			err = p.ExecSql("UPDATE "+myPrefix+"my_dc_transactions SET type_id=?, time = ? WHERE id = ?", cashRequestId, p.BlockData.Time, myId)
			if err != nil {
				return p.ErrInfo(err)
			}
		} else {
			err = p.ExecSql("INSERT INTO "+myPrefix+"my_dc_transactions ( status, type, type_id, to_user_id, amount, currency_id, comment, comment_status ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ? )", "pending", "cash_request", cashRequestId, p.TxMaps.Int64["to_user_id"], p.TxMaps.Money["amount"], p.TxMaps.Int64["currency_id"], utils.BinToHex(p.TxMap["comment"]), "encrypted")
			if err != nil {
				return p.ErrInfo(err)
			}
		}
	}
	p.nfyCashRequest( p.TxMaps.Int64["to_user_id"], &utils.TypeNfyCashRequest{ FromUserId: p.TxUserID, Amount: p.TxMaps.Money["amount"], CurrencyId: p.TxMaps.Int64["currency_id"]} ) 
	return nil
}

func (p *Parser) CashRequestOutRollback() error {
	// возможно нужно обновить таблицу points_status
	err := p.pointsUpdateRollbackMain(p.TxMaps.Int64["to_user_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// обновление нужно, только если данный cash_request единственный с pending, иначе делать пересчет tdc_amount нельзя, т.к. уже были ранее пересчитаны
	cashRequestCount, err := p.Single("SELECT count(id) FROM cash_requests WHERE to_user_id  =  ? AND del_block_id  =  0 AND for_repaid_del_block_id  =  0 AND status  =  'pending'", p.TxMaps.Int64["to_user_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if cashRequestCount == 1 {
		err = p.updPromisedAmountsRollback(p.TxMaps.Int64["to_user_id"], true)
		if err != nil {
			return p.ErrInfo(err)
		}
	} else {
		err = p.updPromisedAmountsCashRequestOutTimeRollback(p.TxMaps.Int64["to_user_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	// при откате учитываем то, что от 1 юзера не может быть более, чем 1 запроса за сутки
	err = p.ExecSql("DELETE FROM cash_requests WHERE time = ? AND from_user_id = ?", p.BlockData.Time, p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.rollbackAI("cash_requests", 1)
	if err != nil {
		return p.ErrInfo(err)
	}

	// отменяем чистку буфера
	err = p.ExecSql("UPDATE wallets_buffer SET del_block_id = 0 WHERE hex(hash) = ?", p.TxMap["hash"])
	if err != nil {
		return p.ErrInfo(err)
	}

	myUserId, _, myPrefix, _, err := p.GetMyUserId(p.TxMaps.Int64["to_user_id"])
	if err != nil {
		return err
	}
	// если наш юзер - получатель запроса
	if p.TxMaps.Int64["to_user_id"] == myUserId {
		affect, err := p.ExecSqlGetAffect("DELETE FROM "+myPrefix+"my_cash_requests WHERE time = ? AND to_user_id = ? AND currency_id = ? AND status = 'pending'", p.BlockData.Time, p.TxMaps.Int64["to_user_id"], p.TxMaps.Int64["currency_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.rollbackAI(myPrefix+"my_cash_requests", affect)
		if err != nil {
			return p.ErrInfo(err)
		}
	} else {
		// наш юзер - отправитель запроса
		myUserId, _, myPrefix, _, err := p.GetMyUserId(p.TxUserID)
		if p.TxUserID == myUserId {
			// обновим статус в нашей локальной табле.
			// у юзера может быть только 1 запрос к 1 юзеру со статусом pending
			err = p.ExecSql("UPDATE "+myPrefix+"my_cash_requests SET status = 'my_pending', cash_request_id = 0 WHERE to_user_id = ? AND status = 'pending'", p.TxMaps.Int64["to_user_id"])
			if err != nil {
				return p.ErrInfo(err)
			}
		}
	}
	p.nfyRollback(p.BlockData.BlockId)

	return nil
}

func (p *Parser) CashRequestOutRollbackFront() error {
	err := p.limitRequestsRollback("cash_requests")
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql("DELETE FROM wallets_buffer WHERE hex(hash) = ?", p.TxMap["hash"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil

}
