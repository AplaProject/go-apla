package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) ChangeCreditorInit() error {

	fields := []map[string]string{{"to_user_id": "int64"}, {"credit_id": "int64"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangeCreditorFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"to_user_id": "bigint", "credit_id": "bigint"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// явлется данный юзер кредитором
	fromUserId, err := p.Single("SELECT from_user_id FROM credits WHERE id  =  ? AND to_user_id  =  ? AND del_block_id  =  0 AND amount > 0", p.TxMaps.Int64["credit_id"], p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if fromUserId == 0 {
		return p.ErrInfo("not a creditor")
	}

	// существет ли полуатель
	err = p.CheckUser(p.TxMaps.Int64["to_user_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	// нельзя давать кредит самому себе или заемщику
	if fromUserId == p.TxMaps.Int64["to_user_id"] || p.TxMaps.Int64["to_user_id"] == p.TxUserID {
		return p.ErrInfo("from_user_id == to_user_id")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["name"], p.TxMap["to_user_id"], p.TxMap["credit_id"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(consts.LIMIT_CHANGE_CREDITOR, "change_creditor", consts.CHANGE_CREDITOR_PERIOD)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangeCreditor() error {
	return p.selectiveLoggingAndUpd([]string{"to_user_id"}, []interface{}{utils.Int64ToStr(p.TxMaps.Int64["to_user_id"])}, "credits", []string{"id"}, []string{utils.Int64ToStr(p.TxMaps.Int64["credit_id"])})
}

func (p *Parser) ChangeCreditorRollback() error {
	return p.selectiveRollback([]string{"to_user_id"}, "credits", "id="+utils.Int64ToStr(p.TxMaps.Int64["credit_id"]), false)
}

func (p *Parser) ChangeCreditorRollbackFront() error {
	return p.limitRequestsRollback("change_creditor")
}
