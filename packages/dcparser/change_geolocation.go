package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) ChangeGeolocationInit() error {

	fields := []map[string]string{{"latitude": "float64"}, {"longitude": "float64"}, {"country": "int64"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangeGeolocationFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"latitude": "coordinate", "longitude": "coordinate", "country": "country"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// является ли данный юзер майнером
	err = p.checkMiner(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	// У юзера не должно быть cash_requests с pending
	err = p.CheckCashRequests(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["latitude"], p.TxMap["longitude"], p.TxMap["country"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(p.Variables.Int64["limit_change_geolocation"], "change_geolocation", p.Variables.Int64["limit_change_geolocation_period"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangeGeolocation() error {

	// возможно нужно обновить таблицу points_status
	err := p.pointsUpdateMain(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.selectiveLoggingAndUpd([]string{"latitude", "longitude", "country"}, []interface{}{p.TxMaps.Float64["latitude"], p.TxMaps.Float64["longitude"], p.TxMaps.Int64["country"]}, "miners_data", []string{"user_id"}, []string{utils.Int64ToStr(p.TxUserID)})
	if err != nil {
		return p.ErrInfo(err)
	}

	// смена местоположения влечет инициацию процедуры выдачи разрешения майнить имеющиеся у юзера валюты в данном местоположении
	// установка promised_amount.status в change_geo возможна, только если до этого был статус mining/pending/change_geo
	// это означает, что нужен пересчет TDC, т.к. до этого момента они майнились
	// логируем предыдущее. Тут ASC, а при откате используем ORDER BY `id` DESC, чтобы не накосячить при уменьшении log_id

	userHolidays, err := p.GetHolidays(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	pointsStatus, err := p.GetPointsStatus(p.TxUserID, p.Variables.Int64["points_update_time"], p.BlockData)
	if err != nil {
		return p.ErrInfo(err)
	}

	rows, err := p.Query(p.FormatQuery(`
				SELECT id,
							 currency_id,
							 status,
							 start_time,
							 amount,
							 tdc_amount,
							 tdc_amount_update,
							 votes_start_time,
							 votes_0,
							 votes_1,
							 log_id
				FROM promised_amount
				WHERE status IN ('mining', 'pending', 'change_geo') AND
							 user_id =? AND
							 currency_id > 1 AND
							 del_block_id = 0 AND
							 del_mining_block_id = 0
				ORDER BY id ASC`), p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id, currency_id, start_time, tdc_amount_update, votes_start_time, votes_0, votes_1, log_id int64
		var status string
		var amount, tdc_amount float64
		err = rows.Scan(&id, &currency_id, &status, &start_time, &amount, &tdc_amount, &tdc_amount_update, &votes_start_time, &votes_0, &votes_1, &log_id)
		if err != nil {
			return p.ErrInfo(err)
		}
		logId, err := p.ExecSqlGetLastInsertId("INSERT INTO log_promised_amount ( status, start_time, tdc_amount, tdc_amount_update, votes_start_time, votes_0, votes_1, block_id, prev_log_id ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ? )", "log_id", status, start_time, tdc_amount, tdc_amount_update, votes_start_time, votes_0, votes_1, p.BlockData.BlockId, log_id)
		if err != nil {
			return p.ErrInfo(err)
		}
		var newTdc float64
		if status == "mining" {
			// то, от чего будем вычислять набежавшие %
			tdcSum := amount + tdc_amount
			// то, что успело набежать
			pct, err := p.GetPct()
			if err != nil {
				return p.ErrInfo(err)
			}
			MaxPromisedAmounts, err := p.GetMaxPromisedAmounts()
			if err != nil {
				return p.ErrInfo(err)
			}
			RepaidAmount, err := p.GetRepaidAmount(currency_id, p.TxUserID)
			if err != nil {
				return p.ErrInfo(err)
			}
			profit, err := p.calcProfit_(tdcSum, tdc_amount_update, p.BlockData.Time, pct[currency_id], pointsStatus, userHolidays, MaxPromisedAmounts[currency_id], currency_id, RepaidAmount)
			if err != nil {
				return p.ErrInfo(err)
			}
			newTdc = tdc_amount + profit
		} else {
			// для статуса 'pending', 'change_geo' нечего пересчитывать, т.к. во время этих статусов ничего не набегает
			newTdc = tdc_amount
		}
		err = p.ExecSql("UPDATE promised_amount SET status = 'change_geo', start_time = 0, tdc_amount = ?, tdc_amount_update = ?, votes_start_time = ?, votes_0 = 0, votes_1 = 0, log_id = ? WHERE id = ?", utils.Round(newTdc, 2), p.BlockData.Time, p.BlockData.Time, logId, id)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	// проверим, не наш ли это user_id
	myUserId, myBlockId, myPrefix, _, err := p.GetMyUserId(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if p.TxUserID == myUserId && myBlockId <= p.BlockData.BlockId {
		err = p.ExecSql("UPDATE " + myPrefix + "my_table SET geolocation_status = 'approved'")
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	return nil
}

func (p *Parser) ChangeGeolocationRollback() error {

	// возможно нужно обновить таблицу points_status
	err := p.pointsUpdateRollbackMain(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.selectiveRollback([]string{"latitude", "longitude", "country"}, "miners_data", "user_id="+utils.Int64ToStr(p.TxUserID), false)
	if err != nil {
		return p.ErrInfo(err)
	}

	// идем в обратном порядке (DESC)
	rows, err := p.Query(p.FormatQuery(`
				SELECT log_id
				FROM promised_amount
				WHERE status = 'change_geo' AND
				             user_id = ? AND
				             del_block_id = 0 AND
				             del_mining_block_id = 0
				ORDER BY id DESC`), p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var log_id int64
		err = rows.Scan(&log_id)
		if err != nil {
			return p.ErrInfo(err)
		}
		// данные, которые восстановим
		logData, err := p.OneRow("SELECT status, start_time, tdc_amount, tdc_amount_update, votes_start_time, votes_0, votes_1, prev_log_id FROM log_promised_amount WHERE log_id  =  ?", log_id).String()
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.ExecSql("UPDATE promised_amount SET status = ?, start_time = ?, tdc_amount = ?, tdc_amount_update = ?, votes_start_time = ?, votes_0 = ?, votes_1 = ?, log_id = ? WHERE log_id = ?", logData["status"], logData["start_time"], logData["tdc_amount"], logData["tdc_amount_update"], logData["votes_start_time"], logData["votes_0"], logData["votes_1"], logData["prev_log_id"], log_id)
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

	// проверим, не наш ли это user_id
	myUserId, _, myPrefix, _, err := p.GetMyUserId(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if p.TxUserID == myUserId {
		err = p.ExecSql("UPDATE " + myPrefix + "my_table SET geolocation_status = 'my_pending'")
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	return nil
}

func (p *Parser) ChangeGeolocationRollbackFront() error {
	return p.limitRequestsRollback("change_geolocation")
}
