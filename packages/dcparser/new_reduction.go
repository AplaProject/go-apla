package dcparser

import (
	"encoding/json"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) NewReductionInit() error {
	var err error
	var fields []map[string]string
	if p.BlockData != nil && p.BlockData.BlockId < 85849 {
		fields = []map[string]string{{"currency_id": "int64"}, {"pct": "string"}, {"sign": "bytes"}}
	} else {
		fields = []map[string]string{{"currency_id": "int64"}, {"pct": "string"}, {"reduction_type": "string"}, {"sign": "bytes"}}
	}
	err = p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}
func (p *Parser) NewReductionFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// является ли данный юзер майнером
	err = p.checkMiner(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"currency_id": "int", "pct": "int"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	if !utils.InSliceInt64(utils.BytesToInt64(p.TxMap["pct"]), consts.ReductionDC) {
		return p.ErrInfo("incorrect pct")
	}

	if p.BlockData != nil && p.BlockData.BlockId < 85849 {
		// для всех тр-ий из старых блоков просто присваем manual, т.к. там не было других типов
		p.TxMaps.String["reduction_type"] = "manual"
	} else {
		verifyData := map[string]string{"reduction_type": "reduction_type"}
		err = p.CheckInputData(verifyData)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	nodePublicKey, err := p.GetNodePublicKey(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(nodePublicKey) == 0 {
		return p.ErrInfo("incorrect user_id")
	}

	currencyId, err := p.CheckCurrencyId(p.TxMaps.Int64["currency_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	if currencyId == 0 {
		return p.ErrInfo("incorrect currency_id")
	}

	forSign := ""
	if p.BlockData != nil && p.BlockData.BlockId < 85849 {
		forSign = fmt.Sprintf("%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["currency_id"], p.TxMap["pct"])
	} else {
		forSign = fmt.Sprintf("%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["currency_id"], p.TxMap["pct"], p.TxMap["reduction_type"])
	}
	CheckSignResult, err := utils.CheckSign([][]byte{nodePublicKey}, forSign, p.TxMap["sign"], true)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if p.TxMaps.String["reduction_type"] == "manual" {
		// проверим, прошло ли 2 недели с момента последнего reduction
		reductionTime, err := p.Single("SELECT max(time) FROM reduction WHERE currency_id  =  ? AND type  =  'manual'", p.TxMaps.Int64["currency_id"]).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		if p.TxTime-reductionTime <= p.Variables.Int64["reduction_period"] {
			return p.ErrInfo("reduction_period error")
		}
	} else {
		reductionTime, err := p.Single("SELECT max(time) FROM reduction WHERE currency_id  =  ? AND type  =  'auto'", p.TxMaps.Int64["currency_id"]).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		// или 48 часов, если это авто-урезание
		if p.TxTime-reductionTime <= consts.AUTO_REDUCTION_PERIOD {
			return p.ErrInfo("reduction_period error")
		}
	}

	if p.TxMaps.String["reduction_type"] == "manual" {

		// получаем кол-во обещанных сумм у разных юзеров по каждой валюте. start_time есть только у тех, у кого статус mining/repaid
		promisedAmount, err := p.DCDB.GetMap(`
					SELECT currency_id, count(user_id) as count
					FROM (
							SELECT currency_id, user_id
							FROM promised_amount
							WHERE start_time < ?  AND
										 del_block_id = 0 AND
										 del_mining_block_id = 0 AND
										 status IN ('mining', 'repaid')
							GROUP BY  user_id, currency_id
							) as t1
					GROUP BY  currency_id`, "currency_id", "count", (p.TxTime - p.Variables.Int64["min_hold_time_promise_amount"]))
		if err != nil {
			return p.ErrInfo(err)
		}
		if len(promisedAmount[utils.Int64ToStr(p.TxMaps.Int64["currency_id"])]) == 0 {
			return p.ErrInfo("empty promised_amount")
		}
		// берем все голоса юзеров по данной валюте
		countVotes, err := p.Single("SELECT count(currency_id) as votes FROM votes_reduction WHERE time > ? AND currency_id  =  ? AND pct  =  ?", (p.TxTime - p.Variables.Int64["reduction_period"]), p.TxMaps.Int64["currency_id"], p.TxMaps.String["pct"]).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		if countVotes < utils.StrToInt64(promisedAmount[utils.Int64ToStr(p.TxMaps.Int64["currency_id"])])/2 {
			return p.ErrInfo("incorrect count_votes")
		}
	} else if p.TxMaps.String["reduction_type"] == "promised_amount" {

		// и недопустимо для WOC
		if p.TxMaps.Int64["currency_id"] == 1 {
			return p.ErrInfo("WOC AUTO_REDUCTION_CASHs")
		}
		// проверим, есть ли хотябы 1000 юзеров, у которых на кошелках есть или была данная валюты
		countUsers, err := p.Single("SELECT count(user_id) FROM wallets WHERE currency_id  =  ?", p.TxMaps.Int64["currency_id"]).Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		if countUsers < consts.AUTO_REDUCTION_PROMISED_AMOUNT_MIN {
			return p.ErrInfo(fmt.Sprintf("AUTO_REDUCTION_PROMISED_AMOUNT_MIN %v < %v", countUsers, consts.AUTO_REDUCTION_PROMISED_AMOUNT_MIN))
		}

		// получаем кол-во DC на кошельках
		sumWallets, err := p.Single("SELECT sum(amount) FROM wallets WHERE currency_id  =  ?", p.TxMaps.Int64["currency_id"]).Float64()
		if err != nil {
			return p.ErrInfo(err)
		}

		// получаем кол-во TDC на обещанных суммах
		sumPromisedAmountTdc, err := p.Single("SELECT sum(tdc_amount) FROM promised_amount WHERE currency_id  =  ?", p.TxMaps.Int64["currency_id"]).Float64()
		if err != nil {
			return p.ErrInfo(err)
		}
		sumWallets += sumPromisedAmountTdc

		// получаем суммы обещанных сумм. при этом не берем те, что имеют просроченные cash_request_out
		sumPromisedAmount, err := p.Single("SELECT sum(amount) FROM promised_amount WHERE status  =  'mining' AND del_block_id  =  0 AND del_mining_block_id  =  0 AND currency_id  =  ? AND (cash_request_out_time  =  0 OR cash_request_out_time > ?)", p.TxMaps.Int64["currency_id"], (p.TxTime - p.Variables.Int64["cash_request_time"])).Float64()
		if err != nil {
			return p.ErrInfo(err)
		}
		log.Debug("sumPromisedAmount", sumPromisedAmount)
		// если обещанных сумм менее чем 100% от объема DC на кошельках, то всё норм, если нет - ошибка
		if sumPromisedAmount >= sumWallets*float64(consts.AUTO_REDUCTION_PROMISED_AMOUNT_PCT) {
			return p.ErrInfo(fmt.Sprintf("error reduction $sum_promised_amount %v >= %v * %v", sumPromisedAmount, sumWallets, consts.AUTO_REDUCTION_PROMISED_AMOUNT_PCT))
		}
	}

	return nil
}

func (p *Parser) NewReduction() error {
	d := (100 - utils.StrToFloat64(p.TxMaps.String["pct"])) / 100
	if utils.StrToFloat64(p.TxMaps.String["pct"]) > 0 {

		// т.к. невозможо 2 отката подряд из-за промежутка в 2 дня между reduction,
		// то можем использовать только бекап на 1 уровень назад вместо _log
		// но для теста полного роллбека до 1-го блока нужно бекапить данные
		data, err := p.DCDB.GetMap("SELECT user_id, amount FROM wallets WHERE currency_id = ?", "user_id", "amount", p.TxMaps.Int64["currency_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		jsonDataWallets, err := json.Marshal(data)
		if err != nil {
			return p.ErrInfo(err)
		}

		err = p.ExecSql("UPDATE wallets SET amount_backup = amount, amount = round(amount*"+utils.Float64ToStr(d)+"+"+utils.Float64ToStr(consts.ROUND_FIX)+", 2) WHERE currency_id = ?", p.TxMaps.Int64["currency_id"])
		if err != nil {
			return p.ErrInfo(err)
		}

		// если бы не урезали amount, то пришлось бы делать пересчет tdc по всем, у кого есть данная валюта
		// после 87826 блока убрано amount_backup = amount, amount = amount*({$d}) т.к. теряется смысл в reduction c type=promised_amount
		data, err = p.DCDB.GetMap("SELECT user_id, tdc_amount FROM promised_amount WHERE currency_id = ?", "user_id", "tdc_amount", p.TxMaps.Int64["currency_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		jsonDataPromisedAmountTdc, err := json.Marshal(data)
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.ExecSql("UPDATE promised_amount SET tdc_amount_backup = tdc_amount, tdc_amount = round(tdc_amount*"+utils.Float64ToStr(d)+"+"+utils.Float64ToStr(consts.ROUND_FIX)+", 2) WHERE currency_id = ?", p.TxMaps.Int64["currency_id"])
		if err != nil {
			return p.ErrInfo(err)
		}

		// все свежие cash_request_out_time отменяем
		data, err = p.DCDB.GetMap("SELECT user_id, cash_request_out_time FROM promised_amount WHERE currency_id = ?", "user_id", "cash_request_out_time", p.TxMaps.Int64["currency_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		jsonDataPromisedAmountTime, err := json.Marshal(data)
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.ExecSql("UPDATE promised_amount SET cash_request_out_time_backup = cash_request_out_time, cash_request_out_time = 0 WHERE currency_id = ? AND cash_request_out_time > ?", p.TxMaps.Int64["currency_id"], (p.BlockData.Time - p.Variables.Int64["cash_request_time"]))
		if err != nil {
			return p.ErrInfo(err)
		}

		// все текущие cash_requests, т.е. по которым не прошло 2 суток
		err = p.ExecSql("UPDATE cash_requests SET del_block_id = ? WHERE currency_id = ? AND status = 'pending' AND time > ?", p.BlockData.BlockId, p.TxMaps.Int64["currency_id"], (p.BlockData.Time - p.Variables.Int64["cash_request_time"]))
		if err != nil {
			return p.ErrInfo(err)
		}

		// форeкс-ордеры
		data, err = p.DCDB.GetMap("SELECT user_id, amount FROM forex_orders WHERE sell_currency_id = ?", "user_id", "amount", p.TxMaps.Int64["currency_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		jsonDataPromisedAmountForex, err := json.Marshal(data)
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.ExecSql("UPDATE forex_orders SET amount_backup = amount, amount = round(amount*"+utils.Float64ToStr(d)+"+"+utils.Float64ToStr(consts.ROUND_FIX)+", 2) WHERE sell_currency_id = ?", p.TxMaps.Int64["currency_id"])
		if err != nil {
			return p.ErrInfo(err)
		}

		// крауд-фандинг
		data, err = p.DCDB.GetMap("SELECT user_id, amount FROM cf_funding WHERE currency_id = ?", "user_id", "amount", p.TxMaps.Int64["currency_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		jsonDataPromisedAmountCF, err := json.Marshal(data)
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.ExecSql("UPDATE cf_funding SET amount_backup = amount, amount = round(amount*"+utils.Float64ToStr(d)+"+"+utils.Float64ToStr(consts.ROUND_FIX)+", 2) WHERE currency_id = ?", p.TxMaps.Int64["currency_id"])
		if err != nil {
			return p.ErrInfo(err)
		}

		log.Debug(`INSERT INTO reduction_backup (block_id, currency_id, cf_funding, forex_orders, promised_amount_cash_request_out_time, promised_amount_tdc_amount, wallets) VALUES (%v, %v, %v, %v, %v, %v, %v) `, p.BlockData.BlockId, p.TxMaps.Int64["currency_id"], jsonDataPromisedAmountCF, jsonDataPromisedAmountForex, jsonDataPromisedAmountTime, jsonDataPromisedAmountTdc, jsonDataWallets)
		err = p.ExecSql(`INSERT INTO reduction_backup (block_id, currency_id, cf_funding, forex_orders, promised_amount_cash_request_out_time, promised_amount_tdc_amount, wallets) VALUES (?, ?, ?, ?, ?, ?, ?)`, p.BlockData.BlockId, p.TxMaps.Int64["currency_id"], jsonDataPromisedAmountCF, jsonDataPromisedAmountForex, jsonDataPromisedAmountTime, jsonDataPromisedAmountTdc, jsonDataWallets)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	rType := ""
	if p.TxMaps.String["reduction_type"] == "manual" {
		rType = "manual"
	} else {
		rType = "auto"
	}
	err := p.ExecSql("INSERT INTO reduction ( time, currency_id, type, pct, block_id ) VALUES ( ?, ?, ?, ?, ? )", p.BlockData.Time, p.TxMaps.Int64["currency_id"], rType, p.TxMaps.String["pct"], p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewReductionRollback() error {

	if utils.StrToFloat64(p.TxMaps.String["pct"]) > 0 {

		jsonData, err := p.OneRow(`SELECT * FROM reduction_backup WHERE block_id = ? AND currency_id = ?`, p.BlockData.BlockId, p.TxMaps.Int64["currency_id"]).Bytes()
		if err != nil {
			return p.ErrInfo(err)
		}
		// крауд-фандинг
		var data map[string]string
		err = json.Unmarshal(jsonData["cf_funding"], &data)
		if err != nil {
			return p.ErrInfo(err)
		}
		for user_id, amount := range data {
			err := p.ExecSql("UPDATE cf_funding SET amount = ? WHERE user_id = ? AND currency_id = ?", amount, user_id, p.TxMaps.Int64["currency_id"])
			if err != nil {
				return p.ErrInfo(err)
			}
		}

		// форекс-ордеры
		data = nil
		err = json.Unmarshal(jsonData["forex_orders"], &data)
		if err != nil {
			return p.ErrInfo(err)
		}
		for user_id, amount := range data {
			err := p.ExecSql("UPDATE forex_orders SET amount = ? WHERE user_id = ? AND sell_currency_id = ?", amount, user_id, p.TxMaps.Int64["currency_id"])
			if err != nil {
				return p.ErrInfo(err)
			}
		}

		// cash_requests del_block_id
		err = p.ExecSql("UPDATE cash_requests SET del_block_id = 0 WHERE del_block_id = ?", p.BlockData.BlockId)
		if err != nil {
			return p.ErrInfo(err)
		}

		// promised_amount cash_request_out_time
		data = nil
		err = json.Unmarshal(jsonData["promised_amount_cash_request_out_time"], &data)
		if err != nil {
			return p.ErrInfo(err)
		}
		for user_id, cash_request_out_time := range data {
			err := p.ExecSql("UPDATE promised_amount SET cash_request_out_time = ? WHERE user_id = ? AND currency_id = ?", cash_request_out_time, user_id, p.TxMaps.Int64["currency_id"])
			if err != nil {
				return p.ErrInfo(err)
			}
		}

		// promised_amount tdc_amount
		data = nil
		err = json.Unmarshal(jsonData["promised_amount_tdc_amount"], &data)
		if err != nil {
			return p.ErrInfo(err)
		}
		for user_id, tdc_amount := range data {
			err := p.ExecSql("UPDATE promised_amount SET tdc_amount = ? WHERE user_id = ? AND currency_id = ?", tdc_amount, user_id, p.TxMaps.Int64["currency_id"])
			if err != nil {
				return p.ErrInfo(err)
			}
		}

		// wallets
		data = nil
		err = json.Unmarshal(jsonData["wallets"], &data)
		if err != nil {
			return p.ErrInfo(err)
		}
		for user_id, amount := range data {
			err := p.ExecSql("UPDATE wallets SET amount = ? WHERE user_id = ? AND currency_id = ?", amount, user_id, p.TxMaps.Int64["currency_id"])
			if err != nil {
				return p.ErrInfo(err)
			}
		}
		/*

			Когда будет много данных, то для свежего отката будем использовать то, что закомменчено

				// крауд-фандинг
				err := p.ExecSql("UPDATE cf_funding SET amount = amount_backup, amount_backup = 0 WHERE currency_id = ?", p.TxMaps.Int64["currency_id"])
				if err != nil {
					return p.ErrInfo(err)
				}
				// форекс-ордеры
				err = p.ExecSql("UPDATE forex_orders SET amount = amount_backup, amount_backup = 0 WHERE sell_currency_id = ?", p.TxMaps.Int64["currency_id"])
				if err != nil {
					return p.ErrInfo(err)
				}
				err = p.ExecSql("UPDATE cash_requests SET del_block_id = 0 WHERE del_block_id = ?", p.BlockData.BlockId)
				if err != nil {
					return p.ErrInfo(err)
				}
				err = p.ExecSql("UPDATE promised_amount SET cash_request_out_time = cash_request_out_time_backup WHERE currency_id = ? AND cash_request_out_time > ?", p.TxMaps.Int64["currency_id"], (p.BlockData.Time - p.Variables.Int64["cash_request_time"]))
				if err != nil {
					return p.ErrInfo(err)
				}
				// после 87826 блока убрано  amount = amount_backup т.к. теряется смысл в reduction c type=promised_amount
				err = p.ExecSql("UPDATE promised_amount SET tdc_amount = tdc_amount_backup WHERE currency_id = ?", p.TxMaps.Int64["currency_id"])
				if err != nil {
					return p.ErrInfo(err)
				}
				err = p.ExecSql("UPDATE wallets SET amount = amount_backup, amount_backup = 0 WHERE currency_id = ?", p.TxMaps.Int64["currency_id"])
				if err != nil {
					return p.ErrInfo(err)
				}
		*/
	}

	affect, err := p.ExecSqlGetAffect("DELETE FROM reduction WHERE block_id = ?", p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	p.rollbackAI("reduction", affect)

	err = p.ExecSql("DELETE FROM reduction_backup WHERE block_id = ?", p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) NewReductionRollbackFront() error {

	return nil
}
