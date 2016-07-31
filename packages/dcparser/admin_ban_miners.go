package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"strings"
)

func (p *Parser) AdminBanMinersInit() error {

	fields := []map[string]string{{"users_ids": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) AdminBanMinersFront() error {

	err := p.generalCheckAdmin()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"users_ids": "users_ids"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	users_ids := strings.Split(p.TxMaps.String["users_ids"], ",")
	for i := 0; i < len(users_ids); i++ {

		// не разжалован ли уже майнер
		status, err := p.Single("SELECT status FROM miners_data WHERE user_id  =  ?", users_ids[i]).String()
		if err != nil {
			return p.ErrInfo(err)
		}
		if status != "miner" {
			return p.ErrInfo("bad miner")
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

func (p *Parser) AdminBanMiners() error {

	users_ids := strings.Split(p.TxMaps.String["users_ids"], ",")
	for i := 0; i < len(users_ids); i++ {
		userId := utils.StrToInt64(users_ids[i])
		// возможно нужно обновить таблицу points_status
		err := p.pointsUpdateMain(userId)

		// переводим майнера из майнеров в юзеры
		data, err := p.OneRow("SELECT miner_id, status, log_id FROM miners_data WHERE user_id  =  ?", userId).String()
		if err != nil {
			return p.ErrInfo(err)
		}

		// логируем текущие значения
		logId, err := p.ExecSqlGetLastInsertId("INSERT INTO log_miners_data ( miner_id, status, block_id, prev_log_id ) VALUES ( ?, ?, ?, ? )", "log_id", data["miner_id"], data["status"], p.BlockData.BlockId, data["log_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.ExecSql("UPDATE miners_data SET status = 'suspended_miner', ban_block_id = ?, miner_id = 0, log_id = ? WHERE user_id = ?", p.BlockData.BlockId, logId, userId)
		if err != nil {
			return p.ErrInfo(err)
		}
		// проверим, не наш ли это user_id
		myUserId, myBlockId, myPrefix, _, err := p.GetMyUserId(userId)
		if err != nil {
			return p.ErrInfo(err)
		}
		if userId == myUserId && myBlockId <= p.BlockData.BlockId {
			err = p.ExecSql("UPDATE " + myPrefix + "my_table SET status = 'user', miner_id = 0, notification_status = 0 WHERE status != 'bad_key'")
			if err != nil {
				return p.ErrInfo(err)
			}
		}

		// изменение статуса юзера влечет смену %, а значит нужен пересчет TDC на обещанных суммах
		// все обещанные суммы, по которым делается превращение tdc->DC
		rows, err := p.Query(p.FormatQuery(`
					SELECT id,
								 amount,
								 currency_id,
								 tdc_amount,
								 tdc_amount_update,
								 start_time,
								 status,
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
		var newTdc float64
		for rows.Next() {
			var id int64
			var amount, currency_id, tdc_amount, tdc_amount_update, start_time, status, log_id string
			err = rows.Scan(&id, &amount, &currency_id, &tdc_amount, &tdc_amount_update, &start_time, &status, &log_id)
			if err != nil {
				return p.ErrInfo(err)
			}
			newTdc, err = p.getTdc(id, userId)
			if err != nil {
				return p.ErrInfo(err)
			}
			addSql := ""
			if status == "repaid" || status == "mining" {
				addSql = fmt.Sprintf("tdcAmount = %f, tdcAmountUpdate = %d, ", utils.Round(newTdc, 2), p.BlockData.Time)
			}

			// логируем текущее значение
			logId, err := p.ExecSqlGetLastInsertId("INSERT INTO log_promised_amount ( tdc_amount, tdc_amount_update, status, block_id, prev_log_id ) VALUES ( ?, ?, ?, ?, ? )", "log_id", tdc_amount, tdc_amount_update, status, p.BlockData.BlockId, log_id)
			if err != nil {
				return p.ErrInfo(err)
			}

			// обновляем TDC и логируем статус
			err = p.ExecSql("UPDATE promised_amount SET "+addSql+" status_backup = status, status = 'suspended', log_id = ? WHERE id = ?", logId, id)
			if err != nil {
				return p.ErrInfo(err)
			}
		}
		p.nfyStatus(userId, `user`)
	}

	return nil
}

func (p *Parser) AdminBanMinersRollback() error {

	users_ids := strings.Split(p.TxMaps.String["users_ids"], ",")
	for i := 0; i < len(users_ids); i++ {

		userId := utils.StrToInt64(users_ids[i])
		// возможно нужно обновить таблицу points_status
		err := p.pointsUpdateRollbackMain(userId)

		// откатываем статус юзера
		logId, err := p.Single("SELECT log_id FROM miners_data WHERE user_id  =  ?", userId).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}

		logData, err := p.OneRow("SELECT status, miner_id, prev_log_id FROM log_miners_data WHERE log_id  =  ?", logId).String()
		if err != nil {
			return p.ErrInfo(err)
		}

		err = p.ExecSql("UPDATE miners_data SET status = ?, miner_id = ?, log_id = ?, ban_block_id = 0 WHERE user_id = ?", logData["status"], logData["miner_id"], logData["prev_log_id"], logId)
		if err != nil {
			return p.ErrInfo(err)
		}

		err = p.ExecSql("UPDATE miners SET active = 1 WHERE miner_id = ?", logData["miner_id"])
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
			err = p.ExecSql("UPDATE "+myPrefix+"my_table SET status = ?, miner_id = ?", logData["status"], logData["miner_id"])
			if err != nil {
				return p.ErrInfo(err)
			}
		}

		// подчищаем _log
		err = p.ExecSql("DELETE FROM log_miners_data WHERE log_id = ?", logId)
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.rollbackAI("log_miners_data", 1)
		if err != nil {
			return p.ErrInfo(err)
		}

		// Откатываем обещанные суммы в обратном прядке
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
			err = p.ExecSql("UPDATE promised_amount SET tdc_amount = ?, tdc_amount_update = ?, status = ?, status_backup = 'null', log_id = ? WHERE id = ?", logData["tdc_amount"], logData["tdc_amount_update"], logData["status"], logData["prev_log_id"], id)
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

func (p *Parser) AdminBanMinersRollbackFront() error {
	return nil
}
