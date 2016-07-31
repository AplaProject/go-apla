package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) DelAutoPaymentInit() error {
	fields := []map[string]string{{"auto_payment_id": "int64"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) DelAutoPaymentFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"auto_payment_id": "bigint"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}
	// проверим, есть ли такой автоплатеж, принадлежащий нашем юзеру
	autoPayment, err := p.Single("SELECT id FROM auto_payments WHERE id  =  ? and sender = ?", p.TxMaps.Int64["auto_payment_id"], p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if autoPayment == 0 {
		return p.ErrInfo("incorrect auto_payment_id")
	}
	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["auto_payment_id"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	return nil
}
func (p *Parser) DelAutoPayment() error {
	err := p.ExecSql("UPDATE auto_payments SET del_block_id = ? WHERE id = ?", p.BlockData.BlockId, p.TxMaps.Int64["auto_payment_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}
func (p *Parser) DelAutoPaymentRollback() error {
	err := p.ExecSql("UPDATE auto_payments SET del_block_id = 0 WHERE id = ?", p.TxMaps.Int64["auto_payment_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return err
}

func (p *Parser) DelAutoPaymentRollbackFront() error {
	return nil
}
