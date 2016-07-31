package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

/* 5
 * Майнер голосует за то, чтобы юзер мог стать или не стать майнером
 * */

func (p *Parser) VotesMinerInit() error {
	fields := []map[string]string{{"vote_id": "int64"}, {"result": "int64"}, {"comment": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) VotesMinerFront() error {
	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	if !utils.CheckInputData(p.TxMap["vote_id"], "bigint") {
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

	// проверим, верно ли указан ID и не закончилось ли голосование
	id, err := p.Single("SELECT id FROM votes_miners WHERE id = ? AND type = 'user_voting' AND votes_end = 0", p.TxMaps.Int64["vote_id"]).Int()
	if err != nil {
		return p.ErrInfo(err)
	}
	if id == 0 {
		return p.ErrInfo("voting is over")
	}

	// проверим, не повторное ли это голосование данного юзера
	num, err := p.Single("SELECT count(user_id) FROM log_votes WHERE user_id = ? AND voting_id = ? AND type = 'votes_miners'", p.TxMaps.Int64["user_id"], p.TxMaps.Int64["vote_id"]).Int()
	if err != nil {
		return p.ErrInfo("double voting")
	}
	if num > 0 {
		return p.ErrInfo("double voting")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["vote_id"], p.TxMap["result"], p.TxMap["comment"])
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

func (p *Parser) VotesMiner() error {
	var notify bool
	// начисляем баллы
	p.points(p.Variables.Int64["miner_points"])

	// обновляем голоса
	err := p.ExecSql("UPDATE votes_miners SET votes_"+utils.Int64ToStr(p.TxMaps.Int64["result"])+" = votes_"+utils.Int64ToStr(p.TxMaps.Int64["result"])+"+1 WHERE id = ?", p.TxMaps.Int64["vote_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	data, err := p.OneRow("SELECT user_id, votes_start_time, votes_0, votes_1 FROM votes_miners WHERE id = ? ", p.TxMaps.Int64["vote_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	// логируем, чтобы юзер {$this->tx_data['user_id']} не смог повторно проголосовать
	err = p.ExecSql("INSERT INTO log_votes ( user_id, voting_id, type ) VALUES ( ?, ?, 'votes_miners' )", p.TxMaps.Int64["user_id"], p.TxMaps.Int64["vote_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	minersData := make(map[string]int64)
	minersData["user_id"] = data["user_id"]
	minersData["votes_start_time"] = data["votes_start_time"]
	minersData["votes_0"] = data["votes_0"]
	minersData["votes_1"] = data["votes_1"]
	minersData["vote_id"] = p.TxMaps.Int64["vote_id"]
	minersData["count_miners"], err = p.Single("SELECT count(miner_id) FROM miners").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	minersData["votes_0_min"] = p.Variables.Int64["miner_votes_0"]
	minersData["votes_1_min"] = p.Variables.Int64["miner_votes_1"]
	minersData["votes_period"] = p.Variables.Int64["miner_votes_period"]

	// -----------------------------------------------------------------------------
	// если голос решающий или голос админа
	// голос админа решающий только при <1000 майнеров.
	// -----------------------------------------------------------------------------
	p.getAdminUserId()
	if p.check24hOrAdminVote(minersData) {
		// перевесили голоса "за" или 1 голос от админа
		if p.checkTrueVotes(minersData) {
			minerId, err := p.insOrUpdMiners(minersData["user_id"])
			if err != nil {
				return p.ErrInfo(err)
			}
			// проверим, не наш ли это user_id
			myUserId, _, myPrefix, _, err := p.GetMyUserId(minersData["user_id"])
			if err != nil {
				return p.ErrInfo(err)
			}
			notify = true
			if minersData["user_id"] == myUserId {
				minerData, err := p.OneRow("SELECT * FROM miners_data WHERE user_id  =  ?", minersData["user_id"]).String()
				if err != nil {
					return p.ErrInfo(err)
				}
				// если это первый http_host, значит майнер не на пуле, а свой хост поднял
				// ставим отметку, чтобы у него автоматом запустился tcp листинг
				count, err := p.Single("SELECT count(user_id) FROM miners_data WHERE http_host  =  ?", minerData["http_host"]).Int()
				if err != nil {
					return p.ErrInfo(err)
				}
				tcpListening := "0"
				if count == 1 {
					tcpListening = "1"
				}
				// обновим статус в нашей локальной табле.
				err = p.ExecSql(`UPDATE `+myPrefix+`my_table
					SET status = 'miner', host_status = 'approved', http_host = ?, tcp_host = ?, face_coords = ?, profile_coords = ?, video_type = ?, video_url_id = ?, miner_id = ?, notification_status = ?, tcp_listening = ?
					WHERE status != 'bad_key'`,
					minerData["http_host"], minerData["tcp_host"], minerData["face_coords"], minerData["profile_coords"], minerData["video_type"], minerData["video_url_id"], minerId, 0, tcpListening)
				if err != nil {
					return p.ErrInfo(err)
				}
			}
		} else { // перевесили голоса "против"
			err = p.ExecSql("UPDATE faces SET status = 'pending' WHERE user_id = ?", data["user_id"])
			if err != nil {
				return p.ErrInfo(err)
			}
		}
		//  ставим "завершено" голосованию
		err = p.ExecSql("UPDATE votes_miners SET votes_end = 1 WHERE id = ?", p.TxMaps.Int64["vote_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		// отметим del_block_id всем, кто голосовал за данного юзера,
		// чтобы через 1440 блоков по крону удалить бесполезные записи
		err = p.ExecSql("UPDATE log_votes SET del_block_id = ? WHERE voting_id = ? AND type = 'votes_miners'", p.BlockData.BlockId, p.TxMaps.Int64["vote_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	// возможно вместе с голосом есть и коммент
	myUserId, _, myPrefix, _, err := p.GetMyUserId(minersData["user_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	if myUserId == minersData["user_id"] {
		err = p.ExecSql("INSERT INTO "+myPrefix+"my_comments ( type, id, comment ) VALUES ( 'miner', ?, ? )", p.TxMaps.Int64["vote_id"], p.TxMaps.String["comment"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	if notify {
		p.nfyStatus(minersData["user_id"], `miner`)
	}
	return nil
}

func (p *Parser) VotesMinerRollback() error {
	userId, err := p.Single("SELECT user_id FROM votes_miners WHERE id  =  ?", p.TxMaps.Int64["vote_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	// обновляем голоса
	err = p.ExecSql("UPDATE votes_miners SET votes_"+utils.Int64ToStr(p.TxMaps.Int64["result"])+" = votes_"+utils.Int64ToStr(p.TxMaps.Int64["result"])+" - 1, votes_end = 0 WHERE id = ?", p.TxMaps.Int64["vote_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// узнаем последствия данного голоса
	data, err := p.OneRow("SELECT miner_id, user_id, status FROM miners_data WHERE user_id  =  ?", userId).String()
	if err != nil {
		return p.ErrInfo(err)
	}

	// удаляем нашу запись из log_votes
	err = p.ExecSql("DELETE FROM log_votes WHERE user_id = ? AND voting_id = ? AND type = 'votes_miners'", p.TxMaps.Int64["user_id"], p.TxMaps.Int64["vote_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// сделал ли голос из юзера майнера?
	if utils.StrToInt(data["miner_id"]) != 0 {
		err = p.insOrUpdMinersRollback(utils.StrToInt64(data["miner_id"]))
		if err != nil {
			return p.ErrInfo(err)
		}

		// меняем статус
		err = p.ExecSql("UPDATE miners_data SET status = 'user', miner_id = 0, reg_time = 0 WHERE user_id = ?", userId)
		if err != nil {
			return p.ErrInfo(err)
		}

		// всем, кому ставили del_block_id, убираем, т.е. отменяем будущее удаление по крону
		err = p.ExecSql("UPDATE log_votes SET del_block_id = 0 WHERE voting_id = ? AND type = 'votes_miners' AND del_block_id = ?", p.TxMaps.Int64["vote_id"], p.BlockData.BlockId)
		if err != nil {
			return p.ErrInfo(err)
		}

		// обновлять faces не нужно, т.к. статус там и так = used

		// проверим, не наш ли это user_id
		myUserId, _, myPrefix, _, err := p.GetMyUserId(utils.StrToInt64(data["user_id"]))
		if err != nil {
			return p.ErrInfo(err)
		}
		if utils.StrToInt64(data["user_id"]) == myUserId {
			// обновим статус в нашей локальной табле.
			// sms/email не трогаем, т.к. смена из-за отката маловажна, и в большинстве случаев статус всё равно сменится.
			err = p.ExecSql("UPDATE " + myPrefix + "my_table SET status = 'user', miner_id = 0, host_status = 'my_pending' WHERE status != 'bad_key'")
			if err != nil {
				return p.ErrInfo(err)
			}
		}
	} else {
		// был ли данный голос решающим-отрицательным
		// т.к. после окончания нодовского голосования и начала юзреского статус у face всегда = used (для избежания одновременной регистрации тысяч клонов), то
		// смена статуса на pending означает, что юзерское голосание было завершено с отрициательным результатом
		err = p.ExecSql("UPDATE faces SET status = 'used' WHERE user_id = ? AND status = 'pending'", userId)
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

func (p *Parser) VotesMinerRollbackFront() error {
	return p.maxDayVotesRollback()
}
