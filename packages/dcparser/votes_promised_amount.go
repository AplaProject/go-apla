package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) VotesPromisedAmountInit() error {
	fields := []map[string]string{{"promised_amount_id": "int64"}, {"result": "int64"}, {"comment": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) VotesPromisedAmountFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"promised_amount_id": "bigint", "result": "vote", "comment": "votes_comment"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// является ли данный юзер майнером
	err = p.checkMiner(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим, не закончилось ли уже голосование и верный ли статус (pending)
	status, err := p.Single("SELECT status FROM promised_amount WHERE id  =  ? AND del_block_id  =  0 AND del_mining_block_id  =  0", p.TxMaps.Int64["promised_amount_id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	if status != "pending" {
		return p.ErrInfo("voting is over")
	}

	// проверим, не повторное ли это голосование данного юзера
	num, err := p.Single("SELECT count(user_id) FROM log_votes WHERE user_id  =  ? AND voting_id  =  ? AND type  =  'promised_amount'", p.TxMaps.Int64["user_id"], p.TxMaps.Int64["promised_amount_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	p.getAdminUserId()
	if num > 0 && p.TxUserID != p.AdminUserId { // админу можно
		return p.ErrInfo("double voting")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["promised_amount_id"], p.TxMap["result"], p.TxMap["comment"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	// лимиты на голоса, чтобы не задосили голосами
	err = p.maxDayVotes()
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) VotesPromisedAmount() error {

	// начисляем баллы
	p.points(p.Variables.Int64["promised_amount_points"])

	// логируем, чтобы юзер {$this->tx_data['user_id']} не смог повторно проголосовать
	err := p.ExecSql("INSERT INTO log_votes ( user_id, voting_id, type ) VALUES ( ?, ?, 'promised_amount' )", p.TxMaps.Int64["user_id"], p.TxMaps.Int64["promised_amount_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// обновляем голоса
	err = p.ExecSql("UPDATE promised_amount SET votes_"+utils.Int64ToStr(p.TxMaps.Int64["result"])+" = votes_"+utils.Int64ToStr(p.TxMaps.Int64["result"])+" + 1 WHERE id = ?", p.TxMaps.Int64["promised_amount_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	promisedAmountData, err := p.OneRow("SELECT log_id, status, start_time, tdc_amount_update, user_id, votes_start_time, votes_0, votes_1 FROM promised_amount WHERE id  =  ?", p.TxMaps.Int64["promised_amount_id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}

	data := make(map[string]int64)
	data["count_miners"], err = p.Single("SELECT count(miner_id) FROM miners").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	data["user_id"] = utils.StrToInt64(promisedAmountData["user_id"])
	data["votes_0"] = utils.StrToInt64(promisedAmountData["votes_0"])
	data["votes_1"] = utils.StrToInt64(promisedAmountData["votes_1"])
	data["votes_start_time"] = utils.StrToInt64(promisedAmountData["votes_start_time"])
	data["votes_0_min"] = p.Variables.Int64["promised_amount_votes_0"]
	data["votes_1_min"] = p.Variables.Int64["promised_amount_votes_1"]
	data["votes_period"] = p.Variables.Int64["promised_amount_votes_period"]

	// -----------------------------------------------------------------------------
	// если голос решающий или голос админа
	// голос админа - решающий только при <1000 майнеров.
	// -----------------------------------------------------------------------------
	err = p.getAdminUserId()
	if err != nil {
		return p.ErrInfo(err)
	}

	if p.check24hOrAdminVote(data) {

		// нужно залогировать, т.к. не известно, какие были status и tdc_amount_update
		logId, err := p.ExecSqlGetLastInsertId("INSERT INTO log_promised_amount ( status, start_time, tdc_amount_update, block_id, prev_log_id ) VALUES ( ?, ?, ?, ?, ? )", "log_id", promisedAmountData["status"], promisedAmountData["start_time"], promisedAmountData["tdc_amount_update"], p.BlockData.BlockId, promisedAmountData["log_id"])
		if err != nil {
			return p.ErrInfo(err)
		}

		// перевесили голоса "за" или 1 голос от админа
		if p.checkTrueVotes(data) {
			err = p.ExecSql("UPDATE promised_amount SET status = 'mining', start_time = ?, tdc_amount_update = ?, log_id = ? WHERE id = ?", p.BlockData.Time, p.BlockData.Time, logId, p.TxMaps.Int64["promised_amount_id"])
			if err != nil {
				return p.ErrInfo(err)
			}
			// есть ли у данного юзера woc
			woc, err := p.Single("SELECT id FROM promised_amount WHERE currency_id  =  1 AND user_id  =  ?", data["user_id"]).Int64()
			if err != nil {
				return p.ErrInfo(err)
			}
			if woc == 0 {
				wocAmount, err := p.Single("SELECT amount FROM max_promised_amounts WHERE id  =  1 ORDER BY time DESC").String()
				if err != nil {
					return p.ErrInfo(err)
				}
				// добавляем WOC
				err = p.ExecSql("INSERT INTO promised_amount ( user_id, amount, currency_id, start_time, status, tdc_amount_update, woc_block_id ) VALUES ( ?, ?, 1, ?, 'mining', ?, ? )", data["user_id"], wocAmount, p.BlockData.Time, p.BlockData.Time, p.BlockData.BlockId)
				if err != nil {
					return p.ErrInfo(err)
				}
			}
		} else { // перевесили голоса "против"
			err = p.ExecSql("UPDATE promised_amount SET status = 'rejected', start_time = 0, tdc_amount_update = ?, log_id = ? WHERE id = ?", p.BlockData.Time, logId, p.TxMaps.Int64["promised_amount_id"])
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
		err = p.ExecSql("INSERT INTO "+myPrefix+"my_comments ( type, id, comment ) VALUES ( 'promised_amount', ?, ? )", p.TxMaps.Int64["promised_amount_id"], p.TxMaps.String["comment"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	return nil
}

func (p *Parser) VotesPromisedAmountRollback() error {

	// вычитаем баллы
	p.pointsRollback(p.Variables.Int64["promised_amount_points"])

	// удаляем логирование, чтобы юзер {$this->tx_data['user_id']} не смог повторно проголосовать
	err := p.ExecSql("DELETE FROM log_votes WHERE user_id = ? AND voting_id = ? AND type = 'promised_amount'", p.TxMaps.Int64["user_id"], p.TxMaps.Int64["promised_amount_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// обновляем голоса
	err = p.ExecSql("UPDATE promised_amount SET votes_"+utils.Int64ToStr(p.TxMaps.Int64["result"])+" = votes_"+utils.Int64ToStr(p.TxMaps.Int64["result"])+" - 1 WHERE id = ?", p.TxMaps.Int64["promised_amount_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	data, err := p.OneRow("SELECT status, user_id, log_id FROM promised_amount WHERE id  =  ?", p.TxMaps.Int64["promised_amount_id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}

	// если статус mining или rejected, значит голос был решающим
	if data["status"] == "mining" || data["status"] == "rejected" {

		// восстановим из лога
		logData, err := p.OneRow("SELECT status, start_time, tdc_amount_update, prev_log_id FROM log_promised_amount WHERE log_id  =  ?", data["log_id"]).String()
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.ExecSql("UPDATE promised_amount SET status = ?, start_time = ?, tdc_amount_update = ?, log_id = ? WHERE id = ?", logData["status"], logData["start_time"], logData["tdc_amount_update"], logData["prev_log_id"], p.TxMaps.Int64["promised_amount_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		// подчищаем _log
		err = p.ExecSql("DELETE FROM log_promised_amount WHERE log_id = ?", data["log_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		p.rollbackAI("log_promised_amount", 1)

		// был ли добавлен woc
		woc, err := p.Single("SELECT id FROM promised_amount WHERE currency_id  =  1 AND woc_block_id  =  ? AND user_id  =  ?", p.BlockData.BlockId, data["user_id"]).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		if woc > 0 {
			err = p.ExecSql("DELETE FROM promised_amount WHERE id = ?", woc)
			if err != nil {
				return p.ErrInfo(err)
			}
			p.rollbackAI("promised_amount", 1)
		}
	}

	return nil
}

func (p *Parser) VotesPromisedAmountRollbackFront() error {
	return p.maxDayVotesRollback()
}
