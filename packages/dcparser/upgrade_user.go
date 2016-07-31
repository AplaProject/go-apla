package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/consts"
)

func (p *Parser) UpgradeUserInit() error {
	fields := []map[string]string{{"sn_type": "string"}, {"sn_url_id": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) UpgradeUserFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"sn_type": "sn_type", "sn_url_id": "sn_url_id"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// юзер может пройти эту процедуру только один раз
	status, err := p.Single("SELECT status FROM users WHERE user_id  =  ?", p.TxMaps.Int64["user_id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	if status == "sn_user" {
		return p.ErrInfo(`status == "sn_user"`)
	}

	attempts, err := p.Single(`SELECT sn_attempts FROM users WHERE user_id = ?`, p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if attempts > consts.SN_USER_ATTEMPTS {
		return p.ErrInfo(`attempts > consts.SN_USER_ATTEMPTS`)
	}

	if p.BlockData == nil || p.BlockData.BlockId > 322674 {
		exists, err := p.Single(`SELECT user_id FROM users WHERE sn_type = ? AND sn_url_id = ? and status != 'rejected_sn_user'`, p.TxMaps.String["sn_type"], p.TxMaps.String["sn_url_id"]).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		if exists > 0 {
			return p.ErrInfo(`exists SN`)
		}
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["sn_type"], p.TxMap["sn_url_id"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(consts.LIMIT_SN_USER, "user_upgrade", consts.LIMIT_SN_USER_PERIOD)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) UpgradeUser() error {
	attempts, err := p.Single(`SELECT sn_attempts FROM users WHERE user_id = ?`, p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	return p.selectiveLoggingAndUpd([]string{"sn_type", "sn_url_id", "votes_start_time", "votes_0", "votes_1", "sn_attempts"}, []interface{}{p.TxMaps.String["sn_type"], p.TxMaps.String["sn_url_id"], p.BlockData.Time, 0, 0, attempts+1}, "users", []string{"user_id"}, []string{utils.Int64ToStr(p.TxUserID)})
}

func (p *Parser) UpgradeUserRollback() error {
	return p.selectiveRollback([]string{"sn_type", "sn_url_id", "votes_start_time", "votes_0", "votes_1", "sn_attempts"}, "users", "user_id="+utils.Int64ToStr(p.TxUserID), false)
}

func (p *Parser) UpgradeUserRollbackFront() error {
	return p.limitRequestsRollback("user_upgrade")
}
