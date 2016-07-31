package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) ChangeKeyActiveInit() error {

	fields := []map[string]string{{"secret": "bytes"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	p.TxMaps.String["secret_hex"] = string(utils.BinToHex(p.TxMaps.Bytes["secret"]))
	if p.TxMaps.String["secret_hex"] == "30" {
		p.TxMaps.Int64["active"] = 0
	} else {
		p.TxMaps.Int64["active"] = 1
	}
	p.TxMaps.String["secret_hex"] = string(utils.BinToHex(p.TxMaps.Bytes["secret"]))

	return nil
}

func (p *Parser) ChangeKeyActiveFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(p.TxMaps.Bytes["secret"]) > 2048 {
		return p.ErrInfo("len secret > 2048")
	}

	// проверим, чтобы не было повторных смен
	changeKey, err := p.Single("SELECT change_key FROM users WHERE user_id  =  ?", p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if changeKey == p.TxMaps.Int64["active"] {
		return p.ErrInfo("active")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMaps.String["secret_hex"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(consts.LIMIT_CHANGE_KEY_ACTIVE, "change_key_active", consts.LIMIT_CHANGE_KEY_ACTIVE_PERIOD)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangeKeyActive() error {
	return p.selectiveLoggingAndUpd([]string{"change_key"}, []interface{}{p.TxMaps.Int64["active"]}, "users", []string{"user_id"}, []string{utils.Int64ToStr(p.TxUserID)})
}

func (p *Parser) ChangeKeyActiveRollback() error {
	return p.selectiveRollback([]string{"change_key"}, "users", "user_id="+utils.Int64ToStr(p.TxUserID), false)
}

func (p *Parser) ChangeKeyActiveRollbackFront() error {
	return p.limitRequestsRollback("change_key_active")
}
