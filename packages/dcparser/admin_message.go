package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) AdminMessageInit() error {

	fields := []map[string]string{{"message": "string"}, {"currency_list": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) AdminMessageFront() error {

	err := p.generalCheckAdmin()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"currency_list": "admin_currency_list"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["message"], p.TxMap["currency_list"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) AdminMessage() error {
	err := p.ExecSql("INSERT INTO alert_messages ( message, currency_list, block_id ) VALUES ( ?, ?, ? )", p.TxMaps.String["message"], p.TxMaps.String["currency_list"], p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) AdminMessageRollback() error {
	affect, err := p.ExecSqlGetAffect("DELETE FROM alert_messages WHERE block_id = ?", p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.rollbackAI("alert_messages", affect)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) AdminMessageRollbackFront() error {
	return nil
}
