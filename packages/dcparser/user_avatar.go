package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) UserAvatarInit() error {

	fields := []map[string]string{{"name": "string"}, {"avatar": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) UserAvatarFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"name": "user_name"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !utils.CheckInputData(p.TxMaps.String["avatar"], "avatar") && p.TxMaps.String["avatar"] != "0" {
		return fmt.Errorf("incorrect avatar")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["name"], p.TxMap["avatar"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(consts.LIMIT_USER_AVATAR, "user_avatar", consts.LIMIT_USER_AVATAR_PERIOD)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) UserAvatar() error {
	return p.selectiveLoggingAndUpd([]string{"name", "avatar"}, []interface{}{p.TxMaps.String["name"], p.TxMaps.String["avatar"]}, "users", []string{"user_id"}, []string{utils.Int64ToStr(p.TxUserID)})
}

func (p *Parser) UserAvatarRollback() error {
	return p.selectiveRollback([]string{"name", "avatar"}, "users", "user_id="+utils.Int64ToStr(p.TxUserID), false)
}

func (p *Parser) UserAvatarRollbackFront() error {
	return p.limitRequestsRollback("user_avatar")
}
