package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"strings"
)

func (p *Parser) AdminUnbanMinersInit() error {

	fields := []map[string]string{{"users_ids": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) AdminUnbanMinersFront() error {

	err := p.generalCheckAdmin()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"users_ids": "users_ids"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим, точно ли были забанены те, кого разбаниваем
	users_ids := strings.Split(p.TxMaps.String["users_ids"], ",")
	for i := 0; i < len(users_ids); i++ {
		num, err := p.Single("SELECT user_id FROM abuses WHERE user_id  =  ?, 'num_rows'", users_ids[i]).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		if num == 0 {
			return p.ErrInfo("no abuses")
		}

		// не разжалован ли уже майнер
		status, err := p.Single("SELECT status FROM miners_data WHERE user_id  =  ?", users_ids[i]).String()
		if err != nil {
			return p.ErrInfo(err)
		}
		if status != "suspended_miner" {
			return p.ErrInfo("status!=suspended_miner")
		}
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["users_ids"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) AdminUnbanMiners() error {

	users_ids := strings.Split(p.TxMaps.String["users_ids"], ",")
	for i := 0; i < len(users_ids); i++ {

		userId := utils.StrToInt64(users_ids[i])

		// возможно нужно обновить таблицу points_status
		err := p.pointsUpdateMain(userId)
		if err != nil {
			return p.ErrInfo(err)
		}
		minerId, err := p.insOrUpdMiners(userId)
		if err != nil {
			return p.ErrInfo(err)
		}

		// проверим, не наш ли это user_id
		myUserId, myBlockId, myPrefix, _, err := p.GetMyUserId(userId)
		if err != nil {
			return p.ErrInfo(err)
		}
		if userId == myUserId && myBlockId <= p.BlockData.BlockId {
			err = p.ExecSql("UPDATE "+myPrefix+"my_table SET status = 'miner', miner_id = ?, notification_status = 0 WHERE status != 'bad_key'", minerId)
			if err != nil {
				return p.ErrInfo(err)
			}
		}

		// изменение статуса юзера влечет обновление tdc_amount_update
		// все обещанные суммы, по которым делается превращение tdc->DC
		rows, err := p.Query(p.FormatQuery(`
					SELECT id,
								 status,
								 status_backup,
								 tdc_amount_update,
								 log_id
					FROM promised_amount
					WHERE user_id = ? AND
								 del_block_id = 0 AND
								 del_mining_block_id = 0
					ORDER BY id ASC`), userId)
		if err != nil {
			return p.ErrInfo(err)
		}
		defer rows.Close()
		for rows.Next() {
			var id, log_id int64
			var status, status_backup, tdc_amount_update string
			err = rows.Scan(&id, &status, &status_backup, &tdc_amount_update, &log_id)
			if err != nil {
				return p.ErrInfo(err)
			}

			logId, err := p.ExecSqlGetLastInsertId("INSERT INTO log_promised_amount ( status, status_backup, block_id, tdc_amount_update, prev_log_id ) VALUES ( ?, ?, ?, ?, ? )", "log_id", status, status_backup, p.BlockData.BlockId, tdc_amount_update, log_id)
			if err != nil {
				return p.ErrInfo(err)
			}
			if log_id > 0 {
				err = p.ExecSql("UPDATE promised_amount SET status = ?, status_backup = 'null', tdc_amount_update = ?, log_id = ? WHERE id = ?", status_backup, p.BlockData.Time, logId, id)
				if err != nil {
					return p.ErrInfo(err)
				}
			} else {
				// если нет log_id, значит promised_amount были добавлены при помощи cash_request_in со статусом suspended уже после того, как было admin_ban_miner
				err = p.ExecSql("UPDATE promised_amount SET status = 'repaid', tdc_amount_update = ? WHERE id = ?", p.BlockData.Time, id)
				if err != nil {
					return p.ErrInfo(err)
				}
			}
		}
		p.nfyStatus(userId, `miner`)

	}

	return nil
}

func (p *Parser) AdminUnbanMinersRollback() error {

	users_ids := strings.Split(p.TxMaps.String["users_ids"], ",")
	for i := 0; i < len(users_ids); i++ {

		userId := utils.StrToInt64(users_ids[i])

		// возможно нужно обновить таблицу points_status
		err := p.pointsUpdateRollbackMain(userId)

		minerId, err := p.Single("SELECT miner_id FROM miners_data WHERE user_id  =  ?", userId).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}

		// откатываем статус юзера
		err = p.ExecSql("UPDATE miners_data SET status = 'suspended_miner', miner_id = 0 WHERE user_id = ?", users_ids[i])
		if err != nil {
			return p.ErrInfo(err)
		}

		err = p.insOrUpdMinersRollback(minerId)
		if err != nil {
			return p.ErrInfo(err)
		}

		// проверим, не наш ли это user_id
		myUserId, _, myPrefix, _, err := p.GetMyUserId(userId)
		if err != nil {
			return p.ErrInfo(err)
		}
		if userId == myUserId {
			// обновим статус в нашей локальной табле.
			// sms/email не трогаем, т.к. скорее всего данные чуть позже вернутся
			err = p.ExecSql("UPDATE " + myPrefix + "my_table SET status = 'suspended_miner', miner_id = 0")
			if err != nil {
				return p.ErrInfo(err)
			}
		}

		// Откатываем обещанные суммы в обратном порядке
		rows, err := p.Query(p.FormatQuery(`
					SELECT id,
								 log_id
					FROM promised_amount
					WHERE user_id = ? AND
								 del_block_id = 0 AND
								 del_mining_block_id = 0
					ORDER BY id DESC`), userId)
		if err != nil {
			return p.ErrInfo(err)
		}
		defer rows.Close()
		for rows.Next() {
			var id, log_id int64
			err = rows.Scan(&id, &log_id)
			if err != nil {
				return p.ErrInfo(err)
			}

			logData, err := p.OneRow("SELECT * FROM log_promised_amount WHERE log_id  =  ?", log_id).String()
			if err != nil {
				return p.ErrInfo(err)
			}

			err = p.ExecSql("UPDATE promised_amount SET status = ?, status_backup = ?, tdc_amount_update = ?, log_id = ? WHERE id = ?", logData["status"], logData["status_backup"], logData["tdc_amount_update"], logData["prev_log_id"], id)
			if err != nil {
				return p.ErrInfo(err)
			}

			// подчищаем _log
			err = p.ExecSql("DELETE FROM log_promised_amount WHERE log_id = ?", log_id)
			if err != nil {
				return p.ErrInfo(err)
			}

			err = p.rollbackAI("log_promised_amount", 1)
			if err != nil {
				return p.ErrInfo(err)
			}
		}
	}
	return nil
}

func (p *Parser) AdminUnbanMinersRollbackFront() error {
	return nil
}
