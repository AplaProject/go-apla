package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) NewAdminInit() error {

	fields := []map[string]string{{"admin_user_id": "int64"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewAdminFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"admin_user_id": "bigint"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// является ли данный юзер майнером
	err = p.checkMiner(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	// является ли юзером новый админ
	err = p.CheckUser(p.TxMaps.Int64["admin_user_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// нодовский ключ
	nodePublicKey, err := p.GetNodePublicKey(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(nodePublicKey) == 0 {
		return p.ErrInfo("incorrect user_id")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["admin_user_id"])
	CheckSignResult, err := utils.CheckSign([][]byte{nodePublicKey}, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	// проверим, прошло ли 2 недели с момента последнего обновления
	adminTime, err := p.Single("SELECT time FROM admin").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if p.TxTime-adminTime <= p.Variables.Int64["new_pct_period"] {
		return p.ErrInfo("14 day error")
	}
	// сколько всего майнеров
	countMiners, err := p.Single("SELECT count(miner_id) FROM miners WHERE active  =  1").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if countMiners < 1000 {
		return p.ErrInfo("countMiners<1000")
	}

	// берем все голоса
	count, err := p.Single("SELECT count(user_id) FROM votes_admin WHERE time > ? AND admin_user_id  =  ?", (p.TxTime - p.Variables.Int64["new_pct_period"]), p.TxMaps.Int64["admin_user_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if count <= countMiners/2 {
		return p.ErrInfo("countMiners")
	}

	err = p.limitRequest(p.Variables.Int64["limit_name"], "name", p.Variables.Int64["limit_name_period"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewAdmin() error {
	return p.selectiveLoggingAndUpd([]string{"user_id", "time"}, []interface{}{p.TxMaps.Int64["admin_user_id"], p.TxTime}, "admin", []string{}, []string{})
}

func (p *Parser) NewAdminRollback() error {
	return p.selectiveRollback([]string{"user_id", "time"}, "admin", "", false)
}

func (p *Parser) NewAdminRollbackFront() error {
	return nil
}
