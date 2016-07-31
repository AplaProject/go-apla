package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) ChangeKeyRequestInit() error {

	fields := []map[string]string{{"to_user_id": "int64"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangeKeyRequestFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"to_user_id": "bigint"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	if p.TxMaps.Int64["to_user_id"] == p.TxUserID {
		return p.ErrInfo("to_user_id == user_id")
	}

	data, err := p.OneRow("SELECT user_id, change_key FROM users WHERE user_id  =  ?", p.TxMaps.Int64["to_user_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	// проверим, есть ли такой юзер
	if len(data) == 0 {
		return p.ErrInfo("!to_user_id")
	}
	// разрешил ли юзер смену ключа админом
	if data["change_key"] == 0 {
		return p.ErrInfo("change_key=0")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["to_user_id"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(consts.LIMIT_CHANGE_KEY_REQUEST, "change_key_request", consts.LIMIT_CHANGE_KEY_REQUEST_PERIOD)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangeKeyRequest() error {
	// change_key_close ставим в 0. чтобы админ через 30 дней мог сменить ключ
	return p.selectiveLoggingAndUpd([]string{"change_key_time", "change_key_close"}, []interface{}{p.BlockData.Time, 0}, "users", []string{"user_id"}, []string{utils.Int64ToStr(p.TxMaps.Int64["to_user_id"])})
}

func (p *Parser) ChangeKeyRequestRollback() error {
	return p.selectiveRollback([]string{"change_key_time", "change_key_close"}, "users", "user_id="+utils.Int64ToStr(p.TxMaps.Int64["to_user_id"]), false)
}

func (p *Parser) ChangeKeyRequestRollbackFront() error {
	return p.limitRequestsRollback("change_key_request")
}
