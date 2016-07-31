package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"encoding/json"
)

func (p *Parser) SwitchPoolInit() error {

	fields := []map[string]string{{"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) SwitchPoolFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// является ли данный юзер майнером
	err = p.checkMiner(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	// получим текущее состояние
	iAmPool, err := p.Single(`
			SELECT i_am_pool
			FROM miners_data
			WHERE user_id = ?`, p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if iAmPool == 0 {
		sumIAmPool, err := p.Single(`SELECT sum(i_am_pool) FROM miners_data`).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		var poolUsers float64
		if sumIAmPool > 0 {
			// проверим лимиты, т.к. хотим стать пулом
			poolUsers, err = p.Single(`SELECT sum(pool_count_users) / sum(i_am_pool) FROM miners_data`).Float64()
			if err != nil {
				return p.ErrInfo(err)
			}
			//current := poolUsers["pool_count_users"]/poolUsers["count"]
			max_pool_users := float64(float64(p.Variables.Int64["max_pool_users"])*0.9)
			// если существующие пулы заняты менее чем на 90%, то новые пулы нам не нужны
			if poolUsers < max_pool_users {
				return p.ErrInfo(fmt.Sprintf("%v < %v", poolUsers, max_pool_users))
			}
		}
	}

	forSign := fmt.Sprintf("%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	// объединяем лимиты для changeHost и для SwitchPool
	err = p.limitRequest(p.Variables.Int64["limit_change_host"], "change_host", p.Variables.Int64["limit_change_host_period"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) SwitchPool() error {
	// получим текущее состояние
	minersData, err := p.OneRow(`
			SELECT i_am_pool, log_id
			FROM miners_data
			WHERE user_id = ?`, p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if minersData["i_am_pool"] == 0 {
		// становимся пулом
		err := p.ExecSql(`UPDATE miners_data SET i_am_pool = 1 WHERE user_id = ?`, p.TxUserID)
		if err != nil {
			return p.ErrInfo(err)
		}
	} else {
		// выключаем режим пула
		users, err := p.GetAll(`SELECT user_id, log_id FROM miners_data WHERE pool_user_id = ?`, -1, p.TxUserID)
		if err != nil {
			return p.ErrInfo(err)
		}
		// сохраним тех, у кого указан данный пул
		jsonData, err := json.Marshal(users)
		if err != nil {
			return p.ErrInfo(err)
		}
		// логируем юзерские данные для роллбека
		logId, err := p.ExecSqlGetLastInsertId(`INSERT INTO log_miners_data (backup_pool_users, prev_log_id) VALUES (?, ?)`, 
												"log_id", string(jsonData), minersData["log_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		// отмечаемся, что мы больше не пул
		err = p.ExecSql(`UPDATE miners_data SET i_am_pool = 0, log_id = ? WHERE user_id = ?`, logId, p.TxUserID);
		if err != nil {
			return p.ErrInfo(err)
		}
		// убираем у юзеров наш пул из pool_user_id
		err = p.ExecSql(`UPDATE miners_data SET pool_user_id = 0, log_id = ? WHERE pool_user_id = ?`, logId, p.TxUserID)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	return nil
}

func (p *Parser) SwitchPoolRollback() error {

	// получим текущее состояние
	minersData, err := p.OneRow(`
			SELECT i_am_pool, log_id
			FROM miners_data
			WHERE user_id = ?`, p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if minersData["i_am_pool"] == 0 {

		// восстановим юзерам, у которых ставили pool_user_id в 0
		log_miners_data, err := p.OneRow(`
				SELECT backup_pool_users, prev_log_id
				FROM log_miners_data
				WHERE log_id = ?`, minersData["log_id"]).Bytes()
		if err != nil {
			return p.ErrInfo(err)
		}

		// включаем режим пула
		err = p.ExecSql(`UPDATE miners_data SET i_am_pool = 1, log_id = ? WHERE user_id = ?`, log_miners_data["prev_log_id"], p.TxUserID)
		if err != nil {
			return p.ErrInfo(err)
		}

		var users []map[string]string
		err = json.Unmarshal(log_miners_data["backup_pool_users"], &users)
		if err != nil {
			return p.ErrInfo(err)
		}

		for _, userData := range users {
			err := p.ExecSql(`UPDATE miners_data SET pool_user_id = ?, log_id = ? WHERE user_id = ?`, p.TxUserID, userData["log_id"], userData["user_id"])
			if err != nil {
				return p.ErrInfo(err)
			}
		}
		err = p.ExecSql(`DELETE FROM log_miners_data WHERE log_id = ?`, minersData["log_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.rollbackAI("log_miners_data", 1)
		if err != nil {
			return p.ErrInfo(err)
		}
	} else {
		// выключаем режим пула
		err := p.ExecSql(`UPDATE miners_data SET i_am_pool = 0 WHERE user_id = ?`, p.TxUserID)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	return nil
}

func (p *Parser) SwitchPoolRollbackFront() error {
	return p.limitRequestsRollback("change_host")
}
