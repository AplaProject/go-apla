package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

/* 31
 * обновляем номер блока photo_block_id и кол-во майнеров photo_max_miner_id,
 * чтобы получить новый набор майнеров,
 * которые должны сохранить фото у себя
 */
func (p *Parser) NewMinerUpdateInit() error {

	fields := []map[string]string{{"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewMinerUpdateFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	forSign := fmt.Sprintf("%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	//  на всякий случай не даем начать нодовское, если идет юзерское голосование
	userVoting, err := p.Single("SELECT id FROM votes_miners WHERE user_id  =  ? AND type  =  'user_voting'", p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if userVoting != 0 {
		return p.ErrInfo("existing user_voting")
	}

	// должна быть запись в miners_data
	userId, err := p.Single("SELECT user_id FROM miners_data WHERE user_id  =  ?", p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if userId == 0 {
		return p.ErrInfo("null miners_data")
	}

	err = p.limitRequest(p.Variables.Int64["limit_votes_miners"], "votes_miners", p.Variables.Int64["limit_votes_miners_period"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewMinerUpdate() error {

	// отменяем голосования по всем предыдущим
	err := p.ExecSql("UPDATE votes_miners SET votes_end = 1, end_block_id = ? WHERE user_id = ? AND type = 'node_voting' AND votes_end = 0", p.BlockData.BlockId, p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	// создаем новое голосование
	err = p.ExecSql("INSERT INTO votes_miners ( type, user_id, votes_start_time ) VALUES ( 'node_voting', ?, ? )", p.TxUserID, p.BlockData.Time)
	if err != nil {
		return p.ErrInfo(err)
	}

	// обновим photo_block_id и photo_max_miner_id чтобы получить
	// 10 новых нодов, которые будут голосовать
	maxMinerId, err := p.Single("SELECT max(miner_id) FROM miners").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.selectiveLoggingAndUpd([]string{"photo_block_id", "photo_max_miner_id"}, []interface{}{p.BlockData.BlockId, maxMinerId}, "miners_data", []string{"user_id"}, []string{utils.Int64ToStr(p.TxUserID)})
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) NewMinerUpdateRollback() error {
	err := p.selectiveRollback([]string{"photo_block_id", "photo_max_miner_id"}, "miners_data", "user_id="+utils.Int64ToStr(p.TxUserID), false)

	// отменяем новое голосование
	err = p.ExecSql("DELETE FROM votes_miners WHERE type = 'node_voting' AND user_id = ? AND votes_start_time = ?", p.TxUserID, p.BlockData.Time)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.rollbackAI("votes_miners", 1)
	if err != nil {
		return p.ErrInfo(err)
	}

	// отменяем отмену голосования
	err = p.ExecSql("UPDATE votes_miners SET votes_end = 0, end_block_id = 0 WHERE user_id = ? AND type = 'node_voting' AND votes_end > 0 AND end_block_id = ?", p.TxUserID, p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) NewMinerUpdateRollbackFront() error {
	return p.limitRequestsRollback("votes_miners")
}
