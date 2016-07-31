package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) AdminAddCurrencyInit() error {

	fields := []map[string]string{{"currency_name": "string"}, {"currency_full_name": "string"}, {"max_promised_amount": "int64"}, {"max_other_currencies": "int64"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) AdminAddCurrencyFront() error {

	err := p.generalCheckAdmin()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"currency_name": "currency_name", "currency_full_name": "currency_full_name", "max_promised_amount": "int", "max_other_currencies": "int"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим, нет ли уже такой валюты
	name, err := p.Single("SELECT name FROM currency WHERE name  =  ?", p.TxMaps.String["currency_name"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(name) > 0 {
		return p.ErrInfo("exists currency_name")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["currency_name"], p.TxMap["currency_full_name"], p.TxMap["max_promised_amount"], p.TxMap["max_other_currencies"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) AdminAddCurrency() error {
	currencyId, err := p.ExecSqlGetLastInsertId("INSERT INTO currency ( name, full_name, max_other_currencies ) VALUES ( ?, ?, ? )", "id", p.TxMaps.String["currency_name"], p.TxMaps.String["currency_full_name"], p.TxMaps.Int64["max_other_currencies"])
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql("INSERT INTO max_promised_amounts ( time, currency_id, amount ) VALUES ( ?, ?, ? )", p.BlockData.Time, currencyId, p.TxMaps.Int64["max_promised_amount"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) AdminAddCurrencyRollback() error {
	currencyId, err := p.Single("SELECT id FROM currency WHERE name  =  ?", p.TxMaps.String["currency_name"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql("DELETE FROM max_promised_amounts WHERE currency_id = ?", currencyId)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.rollbackAI("max_promised_amounts", 1)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql("DELETE FROM currency WHERE id = ?", currencyId)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.rollbackAI("currency", 1)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) AdminAddCurrencyRollbackFront() error {
	return nil
}
