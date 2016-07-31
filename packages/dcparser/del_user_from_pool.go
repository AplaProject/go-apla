package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/consts"
)

func (p *Parser) DelUserFromPoolInit() error {

	fields := []map[string]string{{"del_user_id": "int64"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) DelUserFromPoolFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// является ли данный юзер майнером
	err = p.checkMiner(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	// указан ли у удаляемого юзера наш user_id
	pool_user_id, err := p.Single(`
			SELECT pool_user_id
			FROM miners_data
			WHERE user_id = ?`, p.TxMaps.Int64["del_user_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if pool_user_id != p.TxUserID {
		return p.ErrInfo("pool_user_id != p.TxUserID")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["del_user_id"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(consts.LIMIT_DEL_USER_FROM_POOL, "del_user_from_pool", consts.LIMIT_DEL_USER_FROM_POOL_PERIOD)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) DelUserFromPool() error {

	err := p.selectiveLoggingAndUpd([]string{"pool_user_id"}, []interface{}{0}, "miners_data", []string{"user_id"}, []string{utils.Int64ToStr(p.TxMaps.Int64["del_user_id"])})
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) DelUserFromPoolRollback() error {

	err := p.selectiveRollback([]string{"pool_user_id"}, "miners_data", "user_id="+utils.Int64ToStr(p.TxMaps.Int64["del_user_id"]), false)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) DelUserFromPoolRollbackFront() error {
	return p.limitRequestsRollback("del_user_from_pool")
}
