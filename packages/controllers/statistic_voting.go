package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"

	"fmt"
	"html/template"
	"math"
)

type StatisticVotingPage struct {
	Lang                    map[string]string
	UserId                  int64
	Js                      template.JS
	Divs                    []string
	CurrencyPct             map[int64]map[string]string
	NewPctTpl               map[string]map[string]float64
	PctVotes                map[int64]map[string]map[string]int64
	VotesReferral           map[string][]map[int64]int64
	VotesReduction          map[int64]map[string]string
	PromisedAmount          map[int64]string
	NewMaxOtherCurrencies   map[int64]int64
	MaxOtherCurrenciesVotes map[int64][]map[int64]int64
	MaxPromisedAmountVotes  map[int64][]map[int64]int64
	NewMaxPromisedAmounts   map[int64]int64
	NewReferralPct          map[string]int64
	CurrencyList            map[int64]string
}

func (c *Controller) StatisticVoting() (string, error) {

	var err error

	js := ""
	var divs []string

	/*
	 * Голосование за размер обещанной суммы
	 */
	rows, err := c.Query(c.FormatQuery(`SELECT currency_id, amount, count(user_id) as votes FROM votes_max_promised_amount GROUP BY currency_id, amount`))
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	defer rows.Close()
	maxPromisedAmountVotes := make(map[int64][]map[int64]int64)
	for rows.Next() {
		var currency_id, votes, amount int64
		err = rows.Scan(&currency_id, &amount, &votes)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		maxPromisedAmountVotes[currency_id] = append(maxPromisedAmountVotes[currency_id], map[int64]int64{amount: votes})
	}

	for currencyId, arr := range maxPromisedAmountVotes {
		js += fmt.Sprintf("var max_promised_amounts_%d = [", currencyId)
		for _, data := range arr {
			for k, v := range data {
				js += fmt.Sprintf("[%v, %v],", k, v)
			}
		}
		js = js[:len(js)-1] + "];\n"
		divs = append(divs, fmt.Sprintf("max_promised_amounts_%d", currencyId))
	}

	totalCountCurrencies, err := c.Single("SELECT count(id) FROM currency").Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	newMaxPromisedAmounts := make(map[int64]int64)
	//array []map[int64]int64, min, max, step int64
	for currencyId, data := range maxPromisedAmountVotes {
		newMaxPromisedAmounts[currencyId] = utils.GetMaxVote(data, 0, totalCountCurrencies, 10)
	}

	/*
	 * Голосование за кол-во валют в обещанных суммах
	 */
	rows, err = c.Query(c.FormatQuery(`SELECT currency_id, count, count(user_id) as votes FROM votes_max_other_currencies GROUP BY  currency_id, count`))
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	defer rows.Close()
	maxOtherCurrenciesVotes := make(map[int64][]map[int64]int64)
	for rows.Next() {
		var currency_id, count, votes int64
		err = rows.Scan(&currency_id, &count, &votes)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		maxOtherCurrenciesVotes[currency_id] = append(maxOtherCurrenciesVotes[currency_id], map[int64]int64{count: votes})
	}

	log.Debug("maxOtherCurrenciesVotes", maxOtherCurrenciesVotes)
	newMaxOtherCurrencies := make(map[int64]int64)
	for currencyId, arr := range maxOtherCurrenciesVotes {
		newMaxOtherCurrencies[currencyId] = utils.GetMaxVote(arr, 0, totalCountCurrencies, 10)
		js += fmt.Sprintf("var max_other_currencies_votes_%d = [", currencyId)
		for _, data := range arr {
			log.Debug("data", data)
			for k, v := range data {
				js += fmt.Sprintf("[%v, %v],", k, v)
			}
		}
		js = js[:len(js)-1] + "];\n"
		log.Debug("js", js)
		divs = append(divs, fmt.Sprintf("max_other_currencies_votes_%d", currencyId))
	}

	/*
	 * Голосование за ручное сокращение объема монет
	 * */
	// получаем кол-во обещанных сумм у разных юзеров по каждой валюте. start_time есть только у тех, у кого статус mining/repaid
	promisedAmount_, err := c.GetAll(`
			SELECT currency_id, count(user_id) as count
					FROM (
							SELECT currency_id, user_id
							FROM promised_amount
							WHERE start_time < ? AND
										 del_block_id = 0 AND
										 del_mining_block_id = 0 AND
										 status IN ('mining', 'repaid')
							GROUP BY  user_id, currency_id
							) as t1
					GROUP BY  currency_id
	`, -1, utils.Time()-c.Variables.Int64["min_hold_time_promise_amount"])
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	promisedAmount := make(map[int64]string)
	for _, data := range promisedAmount_ {
		promisedAmount[utils.StrToInt64(data["currency_id"])] = data["count"]
	}

	// берем все голоса юзеров по данной валюте
	votesReduction := make(map[int64]map[string]string)
	votesReduction_, err := c.GetAll(`
			SELECT currency_id,
					  	pct,
					     count(currency_id) as votes
			FROM votes_reduction
			WHERE time > ? AND
						 pct > 0
			GROUP BY currency_id, pct
	`, -1, utils.Time()-c.Variables.Int64["reduction_period"])
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	for _, data := range votesReduction_ {
		curid := utils.StrToInt64(data["currency_id"])
		votesReduction[curid] = make(map[string]string)
		votesReduction[curid][data["pct"]] = data["votes"]
	}

	/*
	 * Голосование за реф. бонусы
	 * */
	refLevels := []string{"first", "second", "third"}
	newReferralPct := make(map[string]int64)
	votesReferral := make(map[string][]map[int64]int64)
	for i := 0; i < len(refLevels); i++ {
		level := refLevels[i]
		// берем все голоса
		votesReferral_, err := c.GetAll(`
				SELECT `+level+`,
							  count(user_id) as votes
				FROM votes_referral
				GROUP BY  `+level+`
		`, -1)
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		for _, data := range votesReferral_ {
			votesReferral[level] = append(votesReferral[level], map[int64]int64{utils.StrToInt64(data[level]): utils.StrToInt64(data["votes"])})
		}

		newReferralPct[level] = utils.GetMaxVote(votesReferral[level], 0, 30, 10)
	}

	for level, arr := range votesReferral {
		js += "var votes_referral_" + level + " = ["
		for _, data := range arr {
			for k, v := range data {
				js += fmt.Sprintf("[%v, %v],", k, v)
			}
		}
		js = js[:len(js)-1] + "];\n"
		divs = append(divs, "votes_referral_"+level)
	}

	/*
	 * Голосоваие за майнеркие и юзерские %
	 * */
	// берем все голоса miner_pct
	pctVotes := make(map[int64]map[string]map[string]int64)
	rows, err = c.Query("SELECT currency_id, pct, count(user_id) as votes FROM votes_miner_pct GROUP BY currency_id, pct ORDER BY currency_id, pct ASC")
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var currency_id, votes int64
		var pct string
		err = rows.Scan(&currency_id, &pct, &votes)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		log.Debug("newpctcurrency_id", currency_id, "pct", pct, "votes", votes)
		if len(pctVotes[currency_id]) == 0 {
			pctVotes[currency_id] = make(map[string]map[string]int64)
		}
		if len(pctVotes[currency_id]["miner_pct"]) == 0 {
			pctVotes[currency_id]["miner_pct"] = make(map[string]int64)
		}
		pctVotes[currency_id]["miner_pct"][pct] = votes
	}

	// берем все голоса user_pct
	rows, err = c.Query("SELECT currency_id, pct, count(user_id) as votes FROM votes_user_pct GROUP BY currency_id, pct ORDER BY currency_id, pct ASC")
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var currency_id, votes int64
		var pct string
		err = rows.Scan(&currency_id, &pct, &votes)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		log.Debug("currency_id", currency_id, "pct", pct, "votes", votes)
		if len(pctVotes[currency_id]) == 0 {
			pctVotes[currency_id] = make(map[string]map[string]int64)
		}
		if len(pctVotes[currency_id]["user_pct"]) == 0 {
			pctVotes[currency_id]["user_pct"] = make(map[string]int64)
		}
		pctVotes[currency_id]["user_pct"][pct] = votes
	}

	log.Debug("pctVotes", pctVotes)

	for currencyId, data := range pctVotes {
		currencyIdStr := utils.Int64ToStr(currencyId)
		divs = append(divs, "miner_pct_"+currencyIdStr)
		divs = append(divs, "user_pct_"+currencyIdStr)

		js += "var miner_pct_" + currencyIdStr + " = ["
		for k, v := range data["miner_pct"] {
			pctY := utils.Round((math.Pow(1+utils.StrToFloat64(k), 3600*24*365)-1)*100, 2)
			js += fmt.Sprintf("[%v, %v],", pctY, v)
		}
		js = js[:len(js)-1] + "];\n"

		js += "var user_pct_" + currencyIdStr + " = ["
		for k, v := range data["user_pct"] {
			pctY := utils.Round((math.Pow(1+utils.StrToFloat64(k), 3600*24*365)-1)*100, 2)
			js += fmt.Sprintf("[%v, %v],", pctY, v)
		}
		js = js[:len(js)-1] + "];\n"
	}

	newPct := make(map[string]map[string]string)
	newPctTpl := make(map[string]map[string]float64)
	var userMaxKey int64
	PctArray := utils.GetPctArray()

	log.Debug("pctVotes", pctVotes)
	for currencyId, data := range pctVotes {

		currencyIdStr := utils.Int64ToStr(currencyId)
		// определяем % для майнеров
		pctArr := utils.MakePctArray(data["miner_pct"])
		log.Debug("currencyIdStr:", currencyIdStr)
		log.Debug("miner_pct:", data["miner_pct"])
		log.Debug("pctArr:", pctArr)
		key := utils.GetMaxVote(pctArr, 0, 390, 100)
		log.Debug("key:", key)
		if len(newPct[currencyIdStr]) == 0 {
			newPct[currencyIdStr] = make(map[string]string)
		}
		newPct[currencyIdStr]["miner_pct"] = utils.GetPctValue(key)
		if len(newPctTpl[currencyIdStr]) == 0 {
			newPctTpl[currencyIdStr] = make(map[string]float64)
		}
		newPctTpl[currencyIdStr]["miner_pct"] = utils.Round((math.Pow(1+utils.StrToFloat64(utils.GetPctValue(key)), 3600*24*365)-1)*100, 2)

		// определяем % для юзеров
		pctArr = utils.MakePctArray(data["user_pct"])
		pctY := utils.ArraySearch(newPct[currencyIdStr]["miner_pct"], PctArray)
		maxUserPctY := utils.Round(utils.StrToFloat64(pctY)/2, 2)
		userMaxKey = utils.FindUserPct(int(maxUserPctY))
		// отрезаем лишнее, т.к. поиск идет ровно до макимального возможного, т.е. до miner_pct/2
		pctArr = utils.DelUserPct(pctArr, userMaxKey)

		key = utils.GetMaxVote(pctArr, 0, userMaxKey, 100)
		newPct[currencyIdStr]["user_pct"] = utils.GetPctValue(key)
		newPctTpl[currencyIdStr]["user_pct"] = utils.Round((math.Pow(1+utils.StrToFloat64(utils.GetPctValue(key)), 3600*24*365)-1)*100, 2)
	}

	log.Debug("newPct", newPct)
	log.Debug("newPctTpl", newPctTpl)

	/*
	 * %/год
	 * */
	currencyPct := make(map[int64]map[string]string)
	for currencyId, name := range c.CurrencyList {
		pct, err := c.OneRow("SELECT * FROM pct WHERE currency_id  =  ? ORDER BY block_id DESC", currencyId).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		currencyPct[currencyId] = make(map[string]string)
		currencyPct[currencyId]["name"] = name
		currencyPct[currencyId]["miner"] = utils.Float64ToStr(utils.Round((math.Pow(1+utils.StrToFloat64(pct["miner"]), 120)-1)*100, 6))
		currencyPct[currencyId]["user"] = utils.Float64ToStr(utils.Round((math.Pow(1+utils.StrToFloat64(pct["user"]), 120)-1)*100, 6))
	}

	TemplateStr, err := makeTemplate("statistic_voting", "statisticVoting", &StatisticVotingPage{
		Lang:                    c.Lang,
		CurrencyList:            c.CurrencyListCf,
		Js:                      template.JS(js),
		Divs:                    divs,
		CurrencyPct:             currencyPct,
		NewPctTpl:               newPctTpl,
		PctVotes:                pctVotes,
		VotesReferral:           votesReferral,
		VotesReduction:          votesReduction,
		PromisedAmount:          promisedAmount,
		NewMaxOtherCurrencies:   newMaxOtherCurrencies,
		MaxOtherCurrenciesVotes: maxOtherCurrenciesVotes,
		MaxPromisedAmountVotes:  maxPromisedAmountVotes,
		NewMaxPromisedAmounts:   newMaxPromisedAmounts,
		NewReferralPct:          newReferralPct,
		UserId:                  c.SessUserId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
