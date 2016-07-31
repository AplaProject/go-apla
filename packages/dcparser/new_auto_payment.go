package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/consts"
)

func (p *Parser) NewAutoPaymentInit() error {
	fields := []map[string]string{{"recipient": "int64"}, {"amount": "money"}, {"commission": "money"}, {"currency_id": "int64"}, {"period": "int64"},  {"comment": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewAutoPaymentFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"amount": "amount", "commission": "amount", "currency_id": "bigint", "period": "bigint", "recipient": "bigint", "comment": "comment"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// существует ли юзер-получатель
	err = p.CheckUser(p.TxMaps.Int64["recipient"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// существует ли валюта
	err = p.CheckUser(p.TxMaps.Int64["currency_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	nodeCommission, err := p.getMyNodeCommission(p.TxMaps.Int64["currency_id"], p.TxUserID, p.TxMaps.Money["amount"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим, удовлетворяет ли нас комиссия, которую предлагает юзер
	if p.TxMaps.Money["commission"] < nodeCommission {
		return p.ErrInfo(fmt.Sprintf("commission %v<%v", p.TxMaps.Money["commission"], nodeCommission))
	}

	if p.TxMaps.Money["amount"] < 0.01 { // 0.01 - минимальная сумма
		return p.ErrInfo("amount")
	}

	if (p.TxMaps.Int64["period"] < 86400 || p.TxMaps.Int64["period"] > 86400*365) {
		return p.ErrInfo("incorrect period")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["recipient"], p.TxMap["amount"], p.TxMap["commission"], p.TxMap["currency_id"], p.TxMap["period"],  p.TxMap["comment"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(consts.LIMIT_AUTO_PAYMENTS, "auto_payments", consts.LIMIT_AUTO_PAYMENTS_PERIOD)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}
func (p *Parser) NewAutoPayment() error {
	err := p.ExecSql(`INSERT INTO auto_payments (amount, currency_id, commission, period, recipient, sender, comment, last_payment_time, block_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			p.TxMaps.Money["amount"], p.TxMaps.Int64["currency_id"], p.TxMaps.Money["commission"],  p.TxMaps.Int64["period"], p.TxMaps.Int64["recipient"], p.TxUserID, p.TxMap["comment"], p.BlockData.Time, p.BlockData.BlockId)
	if err != nil {
		return err
	}
	return nil
}
func (p *Parser) NewAutoPaymentRollback() error {
	//fmt.Println(p.TxMap)
	affect, err := p.ExecSqlGetAffect("DELETE FROM auto_payments WHERE block_id = ? and sender = ?", p.BlockData.BlockId, p.TxUserID)
	if err != nil {
		return utils.ErrInfo(err)
	}

	err = p.rollbackAI("auto_payments", affect)
	if err != nil {
		return utils.ErrInfo(err)
	}
	return err
}

func (p *Parser) NewAutoPaymentRollbackFront() error {
	return p.limitRequestsRollback("auto_payments")
}