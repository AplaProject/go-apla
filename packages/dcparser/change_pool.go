package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) ChangePoolInit() error {
	fields := []map[string]string{{"pool_user_id": "int64"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangePoolFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}
	verifyData := map[string]string{"pool_user_id": "bigint"}

	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// является ли данный юзер майнером
	err = p.checkMiner(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	// TODO: проверить, точно ли это пул
	// есть ли места на выбранном пуле
	count, err := p.Single(`SELECT pool_count_users FROM miners_data WHERE user_id = ?`, p.TxMaps.Int64["pool_user_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if count >= p.Variables.Int64["max_pool_users"] {
		return p.ErrInfo("max_pool_users")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["pool_user_id"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	// общий лимит с change_host
	err = p.limitRequest(p.Variables.Int64["limit_change_host"], "change_host", p.Variables.Int64["limit_change_host_period"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangePool() error {
	err := p.selectiveLoggingAndUpd([]string{"pool_user_id"}, []interface{}{p.TxMaps.Int64["pool_user_id"]}, "miners_data", []string{"user_id"}, []string{utils.Int64ToStr(p.TxUserID)})
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`UPDATE miners_data SET pool_count_users = pool_count_users + 1 WHERE user_id = ?`, p.TxMaps.Int64["pool_user_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangePoolRollback() error {
	err := p.selectiveRollback([]string{"pool_user_id"}, "miners_data", "user_id="+utils.Int64ToStr(p.TxUserID), false)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql(`UPDATE miners_data SET pool_count_users = pool_count_users - 1 WHERE user_id = ?`, p.TxMaps.Int64["pool_user_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangePoolRollbackFront() error {
	return p.limitRequestsRollback("change_host")
}
