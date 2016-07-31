package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	//	"encoding/json"
	//"regexp"
	//"math"
	//	"strings"
	"os"
)

// голосования нодов, которые должны сохранить фото у себя.
// если смог загрузить фото к себе и хэш сошелся - 1, если нет - 0
// эту транзакцию генерит нод со своим ключом

func (p *Parser) VotesNodeNewMinerInit() error {
	fields := []map[string]string{{"vote_id": "int64"}, {"result": "int64"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) VotesNodeNewMinerFront() error {
	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}
	// является ли данный юзер майнером
	err = p.checkMiner(p.TxUserID)
	if err != nil {
		return err
	}

	if !utils.CheckInputData(p.TxMap["result"], "vote") {
		return utils.ErrInfoFmt("incorrect vote")
	}
	// получим public_key
	p.nodePublicKey, err = p.GetNodePublicKey(p.TxUserID)
	if len(p.nodePublicKey) == 0 {
		return utils.ErrInfoFmt("incorrect user_id len(nodePublicKey) = 0")
	}

	if !utils.CheckInputData(p.TxMap["vote_id"], "bigint") {
		return utils.ErrInfoFmt("incorrect bigint")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["vote_id"], p.TxMap["result"])
	CheckSignResult, err := utils.CheckSign([][]byte{p.nodePublicKey}, forSign, p.TxMap["sign"], true)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return utils.ErrInfoFmt("incorrect sign")
	}

	// проверим, верно ли указан ID и не закончилось ли голосование
	id, err := p.Single("SELECT id FROM votes_miners WHERE id = ? AND type = 'node_voting' AND votes_end = 0", p.TxMaps.Int64["vote_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if id == 0 {
		return p.ErrInfo(fmt.Errorf("voting is over"))
	}

	// проверим, не повторное ли это голосование данного юзера
	num, err := p.Single("SELECT count(user_id) FROM log_votes WHERE user_id = ? AND voting_id = ? AND type = 'votes_miners'", p.TxMaps.Int64["user_id"], p.TxMaps.Int64["vote_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if num > 0 {
		return utils.ErrInfoFmt("double voting")
	}

	// нод не должен голосовать более X раз за сутки, чтобы не было доса
	err = p.limitRequest(p.Variables.Int64["node_voting"], "votes_nodes", p.Variables.Int64["node_voting_period"])
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) VotesNodeNewMiner() error {

	var votes [2]int64
	votesData, err := p.OneRow("SELECT user_id, votes_start_time, votes_0, votes_1 FROM votes_miners WHERE id = ?", p.TxMaps.Int64["vote_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	log.Debug("votesData", votesData)
	log.Debug("votesData[user_id]", votesData["user_id"])
	minersData, err := p.OneRow("SELECT photo_block_id, photo_max_miner_id, miners_keepers, pool_user_id, log_id FROM miners_data WHERE user_id = ?", votesData["user_id"]).String()
	log.Debug("minersData", minersData)
	// $votes_data['user_id'] - это юзер, за которого голосуют
	if err != nil {
		return p.ErrInfo(err)
	}

	votes[0] = votesData["votes_0"]
	votes[1] = votesData["votes_1"]
	// прибавим голос
	votes[p.TxMaps.Int64["result"]]++

	// обновляем голоса. При откате просто вычитаем
	err = p.ExecSql("UPDATE votes_miners SET votes_"+utils.Int64ToStr(p.TxMaps.Int64["result"])+" = ? WHERE id = ?", votes[p.TxMaps.Int64["result"]], p.TxMaps.Int64["vote_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// логируем, чтобы юзер {$this->tx_data['user_id']} не смог повторно проголосовать
	err = p.ExecSql("INSERT INTO log_votes (user_id, voting_id, type) VALUES (?, ?, 'votes_miners')", p.TxMaps.Int64["user_id"], p.TxMaps.Int64["vote_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// ID майнеров, у которых сохраняются фотки
	minersIds := utils.GetMinersKeepers(minersData["photo_block_id"], minersData["photo_max_miner_id"], minersData["miners_keepers"], true)

	log.Debug("minersIds", minersIds, len(minersIds))
	// данные для проверки окончания голосования

	minerData := new(MinerData)
	minerData.myMinersIds, err = p.getMyMinersIds()
	if err != nil {
		return p.ErrInfo(err)
	}
	minerData.adminUserId, err = p.GetAdminUserId()
	if err != nil {
		return p.ErrInfo(err)
	}
	minerData.minersIds = minersIds
	minerData.votes0 = votes[0]
	minerData.votes1 = votes[1]
	minerData.minMinersKeepers = p.Variables.Int64["min_miners_keepers"]
	log.Debug("minerData.adminUserId %v", minerData.adminUserId)
	log.Debug("minerData.myMinersIds %v", minerData.myMinersIds)
	log.Debug("minerData.minersIds %v", minerData.minersIds)
	log.Debug("minerData.votes0 %v", minerData.votes0)
	log.Debug("minerData.votes1 %v", minerData.votes1)
	log.Debug("minerData.minMinersKeepers %v", minerData.minMinersKeepers)

	if p.minersCheckVotes1(minerData) || (minerData.votes0 > minerData.minMinersKeepers || int(minerData.votes0) == len(minerData.minersIds)) {
		// отмечаем, что голосование нодов закончено
		err = p.ExecSql("UPDATE votes_miners SET votes_end = 1, end_block_id = ? WHERE id = ?", p.BlockData.BlockId, p.TxMaps.Int64["vote_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	if p.minersCheckVotes1(minerData) || p.minersCheckMyMinerIdAndVotes0(minerData) {
		// отметим del_block_id всем, кто голосовал за данного юзера,
		// чтобы через N блоков по крону удалить бесполезные записи
		err = p.ExecSql("UPDATE log_votes SET del_block_id = ? WHERE voting_id = ? AND type = 'votes_miners'", p.BlockData.BlockId, p.TxMaps.Int64["vote_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	// если набрано >=X голосов "за", то пишем в БД, что юзер готов к проверке людьми
	// либо если набранное кол-во голосов= кол-ву майнеров (актуально в самом начале запуска проекта)
	if p.minersCheckVotes1(minerData) {
		err = p.ExecSql("INSERT INTO votes_miners ( user_id, type, votes_start_time ) VALUES ( ?, 'user_voting', ? )", votesData["user_id"], p.BlockData.Time)
		if err != nil {
			return p.ErrInfo(err)
		}

		// и отмечаем лицо как готовое участвовать в поиске дублей
		err = p.ExecSql("UPDATE faces SET status = 'used' WHERE user_id = ?", votesData["user_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
	} else if p.minersCheckMyMinerIdAndVotes0(minerData) {
		
		// уберем юзера, за которого голосуем из списка прилепленных к пулу
		err = p.ExecSql(`UPDATE miners_data SET pool_count_users = pool_count_users - 1 WHERE user_id = ?`, minersData["pool_user_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.ExecSql(`UPDATE miners_data SET pool_user_id = 0 WHERE user_id = ?`, votesData["user_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		
		// если набрано >5 голосов "против" и мы среди тех X майнеров, которые копировали фото к себе
		// либо если набранное кол-во голосов = кол-ву майнеров (актуально в самом начале запуска проекта)
		facePath := fmt.Sprintf(*utils.Dir+"/public/face_%v.jpg", votesData["user_id"])
		profilePath := fmt.Sprintf(*utils.Dir+"/public/profile_%v.jpg", votesData["user_id"])

		faceRandName := ""
		profileRandName := ""
		// возможно фото к нам не было скопировано, т.к. хост был недоступен.
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			faceRandName = ""
			profileRandName = ""
		} else if _, err := os.Stat(facePath); os.IsNotExist(err) {
			faceRandName = ""
			profileRandName = ""
		} else {
			faceRandName = utils.RandSeq(30)
			profileRandName = utils.RandSeq(30)

			// перемещаем фото в корзину, откуда по крону будем удалять данные
			err = utils.CopyFileContents(facePath, faceRandName)
			if err != nil {
				return p.ErrInfo(err)
			}
			err = os.Remove(facePath)
			if err != nil {
				return p.ErrInfo(err)
			}
			err = utils.CopyFileContents(profilePath, profileRandName)
			if err != nil {
				return p.ErrInfo(err)
			}
			err = os.Remove(profilePath)
			if err != nil {
				return p.ErrInfo(err)
			}

			// если в корзине что-то есть, то логируем
			// отсутствие файлов также логируем, т.к. больше негде, а при откате эти данные очень важны.
			logData, err := p.OneRow("SELECT * FROM recycle_bin WHERE user_id = ?", votesData["user_id"]).String()
			if err != nil {
				return p.ErrInfo(err)
			}
			if len(logData) > 0 {
				logId, err := p.ExecSqlGetLastInsertId("INSERT INTO log_recycle_bin ( user_id, profile_file_name, face_file_name, block_id, prev_log_id ) VALUES ( ?, ?, ?, ?, ? )", "log_id", logData["user_id"], logData["profile_file_name"], logData["face_file_name"], p.BlockData.BlockId, logData["log_id"])
				if err != nil {
					return p.ErrInfo(err)
				}
				err = p.ExecSql("UPDATE recycle_bin SET profile_file_name = ?, face_file_name = ?, log_id = ? WHERE user_id = ?", profileRandName, faceRandName, logId, logData["user_id"])
				if err != nil {
					return p.ErrInfo(err)
				}
			} else {
				err = p.ExecSql("INSERT INTO recycle_bin ( user_id, profile_file_name, face_file_name ) VALUES ( ?, ?, ? )", votesData["user_id"], profileRandName, faceRandName)
				if err != nil {
					return p.ErrInfo(err)
				}
			}
		}
	}

	return nil
}

type MinerData struct {
	adminUserId     int64
	myMinersIds      map[int]int
	minersIds        map[int]int
	votes0           int64
	votes1           int64
	minMinersKeepers int64
}

func (p *Parser) VotesNodeNewMinerRollback() error {

	votesData, err := p.OneRow("SELECT user_id, votes_start_time, votes_0, votes_1 FROM votes_miners WHERE id = ?", p.TxMaps.Int64["vote_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	minersData, err := p.OneRow("SELECT photo_block_id, photo_max_miner_id, miners_keepers, pool_user_id, log_id FROM miners_data WHERE user_id = ?", votesData["user_id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}

	minerData := new(MinerData)
	// запомним голоса- пригодится чуть ниже в minersCheckVotes1
	minerData.votes0 = votesData["votes_0"]
	minerData.votes1 = votesData["votes_1"]

	var votes [2]int64
	votes[0] = votesData["votes_0"]
	votes[1] = votesData["votes_1"]
	// вычтем голос
	votes[p.TxMaps.Int64["result"]]--

	// обновляем голоса
	err = p.ExecSql("UPDATE votes_miners SET votes_"+utils.Int64ToStr(p.TxMaps.Int64["result"])+" = ? WHERE id = ?", votes[p.TxMaps.Int64["result"]], p.TxMaps.Int64["vote_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// удаляем нашу запись из log_votes
	err = p.ExecSql("DELETE FROM log_votes WHERE user_id = ? AND voting_id = ? AND type = 'votes_miners'", p.TxMaps.Int64["user_id"], p.TxMaps.Int64["vote_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	minersIds := utils.GetMinersKeepers(minersData["photo_block_id"], minersData["photo_max_miner_id"], minersData["miners_keepers"], true)
	minerData.myMinersIds, err = p.getMyMinersIds()
	if err != nil {
		return p.ErrInfo(err)
	}
	minerData.minersIds = minersIds
	minerData.minMinersKeepers = p.Variables.Int64["min_miners_keepers"]

	if p.minersCheckVotes1(minerData) || p.minersCheckMyMinerIdAndVotes0(minerData) {

		// отменяем отметку о том, что голосование нодов закончено
		err = p.ExecSql("UPDATE votes_miners SET votes_end = 0, end_block_id = 0 WHERE id = ?", p.TxMaps.Int64["vote_id"])
		if err != nil {
			return p.ErrInfo(err)
		}

		// всем, кому ставили del_block_id, его убираем, т.е. отменяем будущее удаление по крону
		err = p.ExecSql("UPDATE log_votes SET del_block_id = 0 WHERE voting_id = ? AND type = 'votes_miners' AND del_block_id = ? ", p.TxMaps.Int64["vote_id"], p.BlockData.BlockId)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	// если набрано >=5 голосов, то отменяем  в БД, что юзер готов к проверке людьми
	if p.minersCheckVotes1(minerData) {
		// отменяем созданное юзерское голосование
		err = p.ExecSql("DELETE FROM votes_miners WHERE user_id = ? AND votes_start_time = ? AND type = 'user_voting'", votesData["user_id"], p.BlockData.Time)
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.rollbackAI("votes_miners", 1)
		if err != nil {
			return p.ErrInfo(err)
		}

		// и отмечаем лицо как неучаствующее в поиске клонов
		err = p.ExecSql("UPDATE faces SET status = 'pending' WHERE user_id = ?", votesData["user_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
	} else if p.minersCheckMyMinerIdAndVotes0(minerData) {

		// вернем юзера, за которого голосуем в список прилепленных к пулу
		err = p.ExecSql(`UPDATE miners_data SET pool_count_users = pool_count_users + 1 WHERE user_id = ?`, minersData["pool_user_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.ExecSql(`UPDATE miners_data SET pool_user_id = ? WHERE user_id = ?`, minersData["pool_user_id"], votesData["user_id"])
		if err != nil {
			return p.ErrInfo(err)
		}

		// если фото плохое и мы среди тех 10 майнеров, которые копировали (или нет) фото к себе,
		// а затем переместили фото в корзину

		// получаем rand_name из логов
		data, err := p.OneRow("SELECT profile_file_name, face_file_name FROM recycle_bin WHERE user_id = ?", votesData["user_id"]).String()
		if err != nil {
			return p.ErrInfo(err)
		}

		// перемещаем фото из корзины, если есть, что перемещать
		if len(data["profile_file_name"]) > 0 && len(data["face_file_name"]) > 0 {
			utils.CopyFileContents("recycle_bin/"+data["face_file_name"], *utils.Dir+"/public/face_"+utils.Int64ToStr(votesData["user_id"])+".jpg")
			utils.CopyFileContents("recycle_bin/"+data["profile_file_name"], *utils.Dir+"/public/profile_"+utils.Int64ToStr(votesData["user_id"])+".jpg")
		}
		p.generalRollback("recycle_bin", votesData["user_id"], "", false)
	}

	return nil
}

func (p *Parser) VotesNodeNewMinerRollbackFront() error {
	return p.limitRequestsRollback("votes_nodes")
}
