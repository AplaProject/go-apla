package dcparser

import (
	"database/sql"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/consts"
)

func (p *Parser) ChangePromisedAmountInit() error {
	fields := []map[string]string{{"promised_amount_id": "int64"}, {"amount": "money"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangePromisedAmountFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"promised_amount_id": "int", "amount": "amount"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// юзер должен быть или miner, или passive_miner, т.е. иметь miner_id. не даем майнерам, которых забанил админ, добавлять новые обещанные суммы.
	err = p.checkMiner(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	// верный ли id. менять сумму можно, только когда статус mining
	// нельзя изменить woc (currency_id=1)
	promisedAmountData, err := p.OneRow("SELECT id, currency_id FROM promised_amount WHERE id  =  ? AND status  =  'mining' AND currency_id > 1 AND del_block_id  =  0 AND del_mining_block_id  =  0", p.TxMaps.Int64["promised_amount_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if promisedAmountData["id"] == 0 {
		return p.ErrInfo("incorrect promised_amount_id")
	}

	maxPromisedAmount, err := p.GetMaxPromisedAmount(promisedAmountData["currency_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	
	// пока нет хотя бы 1000 майнеров по этой валюте, ограничиваем размер обещанной суммы
	countMiners, err := p.Single("SELECT count(id) FROM promised_amount where currency_id = ? AND status='mining'", promisedAmountData["currency_id"]).Int64()

	if countMiners < 1000 && (p.BlockData == nil || p.BlockData.BlockId > 297496) {
		maxPromisedAmount = float64(consts.MaxGreen[promisedAmountData["currency_id"]])
	}
	
	// т.к. можно перевести из mining в repaid, где нет лимитов, и так проделать много раз, то
	// нужно жестко лимитировать ОБЩУЮ сумму по всем promised_amount данной валюты
	repaidAmount, err := p.GetRepaidAmount(promisedAmountData["currency_id"], p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if p.TxMaps.Money["amount"]+repaidAmount > maxPromisedAmount {
		return p.ErrInfo("incorrect amount")
	}

	err = p.CheckCashRequests(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["promised_amount_id"], p.TxMap["amount"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(p.Variables.Int64["limit_promised_amount"], "promised_amount", p.Variables.Int64["limit_promised_amount_period"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangePromisedAmount() error {

	// возможно нужно обновить таблицу points_status
	err := p.pointsUpdateMain(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	pct, err := p.GetPct()
	if err != nil {
		return p.ErrInfo(err)
	}

	maxPromisedAmounts, err := p.GetMaxPromisedAmounts()
	if err != nil {
		return p.ErrInfo(err)
	}

	// логируем предыдущее
	var prevLogId, currencyId, tdcAmountUpdate int64
	var amount, tdcAmount float64
	err = p.QueryRow(p.FormatQuery("SELECT log_id, currency_id, amount, tdc_amount, tdc_amount_update FROM promised_amount WHERE id  =  ?"), p.TxMaps.Int64["promised_amount_id"]).Scan(&prevLogId, &currencyId, &amount, &tdcAmount, &tdcAmountUpdate)
	if err != nil && err != sql.ErrNoRows {
		return p.ErrInfo(err)
	}

	logId, err := p.ExecSqlGetLastInsertId("INSERT INTO log_promised_amount ( amount, tdc_amount, tdc_amount_update, block_id, prev_log_id ) VALUES ( ?, ?, ?, ?, ? )", "log_id", amount, tdcAmount, tdcAmountUpdate, p.BlockData.BlockId, prevLogId)
	if err != nil {
		return p.ErrInfo(err)
	}

	userHolidays, err := p.GetHolidays(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	pointsStatus, err := p.GetPointsStatus(p.TxUserID, p.Variables.Int64["points_update_time"], p.BlockData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// то, от чего будем вычислять набежавшие %
	tdcSum := amount + tdcAmount

	// то, что успело набежать
	repaidAmount, err := p.GetRepaidAmount(currencyId, p.TxUserID)
	calcProfit, err := p.calcProfit_(tdcSum, tdcAmountUpdate, p.BlockData.Time, pct[currencyId], pointsStatus, userHolidays, maxPromisedAmounts[currencyId], currencyId, repaidAmount)
	newTdc := tdcAmount + calcProfit
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql("UPDATE promised_amount SET amount = ?, tdc_amount = ?, tdc_amount_update = ?, log_id = ? WHERE id = ?", p.TxMaps.Money["amount"], utils.Round(newTdc, 2), p.BlockData.Time, logId, p.TxMaps.Int64["promised_amount_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) ChangePromisedAmountRollback() error {
	// возможно нужно обновить таблицу points_status
	err := p.pointsUpdateRollbackMain(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	logId, err := p.Single("SELECT log_id FROM promised_amount WHERE id  =  ?", p.TxMaps.Int64["promised_amount_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	// данные, которые восстановим
	logData, err := p.OneRow("SELECT amount, tdc_amount, tdc_amount_update, prev_log_id FROM log_promised_amount WHERE log_id  =  ?", logId).String()
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql("UPDATE promised_amount SET amount = ?, tdc_amount = ?, tdc_amount_update = ?, log_id = ? WHERE id = ?", logData["amount"], logData["tdc_amount"], logData["tdc_amount_update"], logData["prev_log_id"], p.TxMaps.Int64["promised_amount_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// подчищаем _log
	err = p.ExecSql("DELETE FROM log_promised_amount WHERE log_id = ?", logId)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.rollbackAI("log_promised_amount", 1)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) ChangePromisedAmountRollbackFront() error {
	return p.limitRequestsRollback("promised_amount")
}
