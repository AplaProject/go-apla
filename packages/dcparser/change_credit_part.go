package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) ChangeCreditPartInit() error {

	fields := []map[string]string{{"pct": "float64"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangeCreditPartFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"pct": "credit_pct"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	if p.TxMaps.Float64["pct"] > 100 {
		return p.ErrInfo("incorrect pct")
	}

	creditPart, err := p.Single("SELECT credit_part FROM users WHERE user_id  =  ?", p.TxUserID).Float64()
	if err != nil {
		return p.ErrInfo(err)
	}
	// проверим, есть ли активные кредиты
	credits, err := p.Single("SELECT id FROM credits WHERE from_user_id  =  ? AND del_block_id  =  0 AND amount > 0", p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	// нельзя увеличивать credit_part, т.к. это будет нечестно по отношению к кредиторам
	if p.TxMaps.Float64["pct"] > creditPart && credits > 0 {
		return p.ErrInfo("incorrect pct")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["pct"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(consts.LIMIT_CHANGE_CREDIT_PART, "change_credit_part", consts.LIMIT_CHANGE_CREDIT_PART_PERIOD)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangeCreditPart() error {
	return p.selectiveLoggingAndUpd([]string{"credit_part"}, []interface{}{utils.Float64ToStr(p.TxMaps.Float64["pct"])}, "users", []string{"user_id"}, []string{utils.Int64ToStr(p.TxUserID)})
}

func (p *Parser) ChangeCreditPartRollback() error {
	return p.selectiveRollback([]string{"credit_part"}, "users", "user_id="+utils.Int64ToStr(p.TxUserID), false)

}

func (p *Parser) ChangeCreditPartRollbackFront() error {
	return p.limitRequestsRollback("change_credit_part")
}
