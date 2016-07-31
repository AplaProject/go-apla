package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/consts"
)

func (p *Parser) VotesSnUserInit() error {
	fields := []map[string]string{{"sn_user_id": "int64"}, {"result": "int64"}, {"comment": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) VotesSnUserFront() error {
	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	if !utils.CheckInputData(p.TxMap["sn_user_id"], "bigint") {
		return p.ErrInfo("incorrect vote_id")
	}
	if !utils.CheckInputData(p.TxMap["result"], "vote") {
		return p.ErrInfo("incorrect vote_id")
	}
	if !utils.CheckInputData(p.TxMap["comment"], "votes_comment") {
		return p.ErrInfo("incorrect comment")
	}

	// является ли данный юзер майнером
	err = p.checkMiner(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим, не закончилось ли уже голосование и верный ли статус
	status, err := p.Single("SELECT status FROM users WHERE user_id  =  ?", p.TxMaps.Int64["sn_user_id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	if status != "user" {
		return p.ErrInfo("voting is over")
	}

	// проверим, не повторное ли это голосование данного юзера
	num, err := p.Single("SELECT count(user_id) FROM log_votes WHERE user_id  =  ? AND voting_id  =  ? AND type  =  'sn_user'", p.TxMaps.Int64["user_id"], p.TxMaps.Int64["sn_user_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	p.getAdminUserId()
	if num > 0 && p.TxUserID != p.AdminUserId { // админу можно
		return p.ErrInfo("double voting")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["sn_user_id"], p.TxMap["result"], p.TxMap["comment"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	// защита от доса
	err = p.maxDayVotes()
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) VotesSnUser() error {
	var  notify bool
	// начисляем баллы
	p.points(p.Variables.Int64["miner_points"])
	// логируем, чтобы юзер {$this->tx_data['user_id']} не смог повторно проголосовать
	err := p.ExecSql("INSERT INTO log_votes ( user_id, voting_id, type ) VALUES ( ?, ?, 'sn_user' )", p.TxUserID, p.TxMaps.Int64["sn_user_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// обновляем голоса
	err = p.ExecSql("UPDATE users SET votes_"+utils.Int64ToStr(p.TxMaps.Int64["result"])+" = votes_"+utils.Int64ToStr(p.TxMaps.Int64["result"])+" + 1 WHERE user_id = ?", p.TxMaps.Int64["sn_user_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	userData, err := p.OneRow("SELECT log_id, status, user_id, votes_start_time, votes_0, votes_1 FROM users WHERE user_id  =  ?", p.TxMaps.Int64["sn_user_id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}

	data := make(map[string]int64)
	data["count_miners"], err = p.Single("SELECT count(miner_id) FROM miners").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	data["user_id"] = utils.StrToInt64(userData["user_id"])
	data["votes_0"] = utils.StrToInt64(userData["votes_0"])
	data["votes_1"] = utils.StrToInt64(userData["votes_1"])
	data["votes_start_time"] = utils.StrToInt64(userData["votes_start_time"])
	data["votes_0_min"] = consts.LIMIT_SN_VOTES_0
	data["votes_1_min"] = consts.LIMIT_SN_VOTES_1
	data["votes_period"] = consts.LIMIT_SN_VOTES_PERIOD

	// -----------------------------------------------------------------------------
	// если голос решающий или голос админа
	// голос админа - решающий только при <1000 майнеров.
	// -----------------------------------------------------------------------------
	err = p.getAdminUserId()
	if err != nil {
		return p.ErrInfo(err)
	}

	if p.check24hOrAdminVote(data) {

		// перевесили голоса "за" или 1 голос от админа
		if p.checkTrueVotes(data) {
			err = p.selectiveLoggingAndUpd([]string{"status"}, []interface{}{"sn_user"}, "users", []string{"user_id"}, []string{utils.Int64ToStr(p.TxMaps.Int64["sn_user_id"])})
			notify = true
			if err != nil {
				return p.ErrInfo(err)
			}
		} else { // перевесили голоса "против"
			err = p.selectiveLoggingAndUpd([]string{"status"}, []interface{}{"rejected_sn_user"}, "users", []string{"user_id"}, []string{utils.Int64ToStr(p.TxMaps.Int64["sn_user_id"])})
			if err != nil {
				return p.ErrInfo(err)
			}
		}
	}

	// возможно с голосом пришел коммент
	myUserId, _, myPrefix, _, err := p.GetMyUserId(p.TxMaps.Int64["user_id"])
	if err != nil {
		return err
	}
	if p.TxUserID == myUserId {
		err = p.ExecSql("INSERT INTO "+myPrefix+"my_comments ( type, id, comment ) VALUES ( 'sn_user', ?, ? )", p.TxMaps.Int64["sn_user_id"], p.TxMaps.String["comment"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	if notify {
		p.nfyStatus(data["user_id"], `sn_user`)
	}
	return nil
}

func (p *Parser) VotesSnUserRollback() error {

	// удаляем логирование, чтобы юзер {$this->tx_data['user_id']} не смог повторно проголосовать
	err := p.ExecSql("DELETE FROM log_votes WHERE user_id = ? AND voting_id = ? AND type = 'sn_user'", p.TxUserID, p.TxMaps.Int64["sn_user_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	status, err := p.Single(`SELECT status FROM users WHERE user_id = ?`, p.TxMaps.Int64["sn_user_id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}

	// если голос был решающим
	if status != "user" {
		err = p.selectiveRollback([]string{"status"}, "users", "user_id="+utils.Int64ToStr(p.TxMaps.Int64["sn_user_id"]), false)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	// вычитаем баллы
	err = p.pointsRollback(p.Variables.Int64["miner_points"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) VotesSnUserRollbackFront() error {
	return p.maxDayVotesRollback()
}
