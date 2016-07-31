package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)


func (p *Parser) NewRestrictedPromisedAmountInit() error {
	fields := []map[string]string{{"currency_id": "int64"}, {"amount": "money"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewRestrictedPromisedAmountFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"currency_id": "int", "amount": "amount"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим, существует ли такая валюта
	if ok, err := p.CheckCurrency(p.TxMaps.Int64["currency_id"]); !ok {
		return p.ErrInfo(err)
	}

	// проверим нет ли у юзера других сумм restricted
	promised_amount_restricted, err := p.Single("SELECT id FROM promised_amount_restricted WHERE user_id = ?", p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if promised_amount_restricted > 0 {
		return p.ErrInfo("exists promised_amount_restricted")
	}

	/*Даем юзера вначале увидеть его монеты, а перед их переводом на счет просим указать акк соц. сети
	if p.BlockData == nil || p.BlockData.BlockId > 310000 {
		// прошел ли проверку соц. сети
		status, err := p.Single("SELECT status FROM users WHERE user_id = ?", p.TxUserID).String()
		if err != nil {
			return p.ErrInfo(err)
		}
		if status != "sn_user" {
			return p.ErrInfo("!sn_user")
		}
	}*/

	if p.TxMaps.Float64["amount"] > 30 || p.TxMaps.Int64["currency_id"] != 72 {
		return p.ErrInfo("incorrect amount currency_id")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["currency_id"], p.TxMap["amount"])

	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) NewRestrictedPromisedAmount() error {

	//добавляем promised_amount_restricted в БД
	err := p.ExecSql(`
				INSERT INTO promised_amount_restricted (
						user_id,
						amount,
						currency_id,
						last_update
					)
					VALUES (
						` + utils.Int64ToStr(p.TxMaps.Int64["user_id"]) + `,
						` + utils.Float64ToStr(p.TxMaps.Money["amount"]) + `,
						` + utils.Int64ToStr(p.TxMaps.Int64["currency_id"]) + `,
						` + utils.Int64ToStr(p.BlockData.Time) + `
					)`)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) NewRestrictedPromisedAmountRollback() error {
	err := p.ExecSql("DELETE FROM promised_amount_restricted WHERE user_id = ?", p.TxMaps.Int64["user_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.rollbackAI("promised_amount_restricted", 1)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewRestrictedPromisedAmountRollbackFront() error {
	return nil
}
