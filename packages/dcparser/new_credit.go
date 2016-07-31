package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) NewCreditInit() error {

	fields := []map[string]string{{"to_user_id": "int64"}, {"amount": "money"}, {"currency_id": "int64"}, {"pct": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewCreditFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"to_user_id": "bigint", "amount": "amount", "pct": "credit_pct"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}
	if p.TxMaps.Money["amount"] < 0.01 {
		return p.ErrInfo("incorrect amount")
	}

	// является ли данный юзер майнером
	/*err = p.checkMiner(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}*/
	// нельзя давать кредит самому себе
	if p.TxMaps.Int64["user_id"] == p.TxMaps.Int64["to_user_id"] {
		return p.ErrInfo("user_id = to_user_id")
	}

	err = p.CheckUser(p.TxMaps.Int64["to_user_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим, существует ли такая валюта в таблице DC-валют
	if ok, _ := p.CheckCurrency(p.TxMaps.Int64["currency_id"]); !ok {
		if ok, err := p.CheckCurrencyCF(p.TxMaps.Int64["currency_id"]); !ok {
			return p.ErrInfo(err)
		}
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["to_user_id"], p.TxMap["amount"], p.TxMap["currency_id"], p.TxMap["pct"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	admin, err := p.GetAdminUserId()
	if err != nil {
		return p.ErrInfo(err)
	}
	if p.TxUserID == admin {
		err = p.limitRequest(500, "new_credit", consts.NEW_CREDIT_PERIOD)
	} else {
		err = p.limitRequest(consts.LIMIT_NEW_CREDIT, "new_credit", consts.NEW_CREDIT_PERIOD)
	}
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewCredit() error {
	return p.ExecSql("INSERT INTO credits ( time, amount, from_user_id, to_user_id, currency_id, pct, tx_hash, tx_block_id ) VALUES ( ?, ?, ?, ?, ?, ?, [hex], ? )", p.BlockData.Time, p.TxMaps.Money["amount"], p.TxUserID, p.TxMaps.Int64["to_user_id"], p.TxMaps.Int64["currency_id"], p.TxMaps.String["pct"], p.TxHash, p.BlockData.BlockId)
}

func (p *Parser) NewCreditRollback() error {
	log.Debug("p.TxHash %s", p.TxHash)
	log.Debug("p.BlockData.BlockId %v", p.BlockData.BlockId)
	affect, err := p.ExecSqlGetAffect("DELETE FROM credits WHERE tx_block_id = ? AND hex(tx_hash) = ?", p.BlockData.BlockId, p.TxHash)
	if err != nil {
		return p.ErrInfo(err)
	}
	log.Debug("affect %v", affect)
	err = p.rollbackAI("credits", 1)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewCreditRollbackFront() error {
	return p.limitRequestsRollback("new_credit")
}
