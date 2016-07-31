package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"math"
	"time"
)

type StatisticPage struct {
	Lang                       map[string]string
	UserId                     int64
	UserInfoId                 int64
	CurrencyList               map[int64]string
	SumWallets                 map[int64]float64
	MaxAmount                  map[int64]float64
	SumPromisedAmount          map[string]string
	PromisedAmountMiners       map[string]string
	WalletsUsers               map[string]string
	UserInfoWallets            []utils.DCAmounts
	Credits                    map[string]string
	PromisedAmountListAccepted []utils.PromisedAmounts
	CountUsers                 int64
	CurrencyPct                map[int64]map[string]string
	Reduction                  []map[string]string
}

func (c *Controller) Statistic() (string, error) {

	var err error

	sumWallets := make(map[int64]float64)
	// получаем кол-во DC на кошельках
	rows, err := c.Query(`
			SELECT currency_id,
					     sum(amount) as sum_amount
			FROM wallets
			GROUP BY currency_id
			`)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var currency_id int64
		var sum_amount float64
		err = rows.Scan(&currency_id, &sum_amount)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		sumWallets[currency_id] = sum_amount
	}
	// получаем кол-во TDC на обещанных суммах
	rows, err = c.Query(`
			SELECT currency_id,
			  		     sum(tdc_amount) as sum_amount
			FROM promised_amount
			GROUP BY currency_id
			`)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var currency_id int64
		var sum_amount float64
		err = rows.Scan(&currency_id, &sum_amount)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if sumWallets[currency_id] > 0 {
			sumWallets[currency_id] += sum_amount
		} else {
			sumWallets[currency_id] = sum_amount
		}
	}

	// получаем суммы обещанных сумм
	sumPromisedAmount, err := c.GetMap(`
			SELECT currency_id,
						sum(amount) as sum_amount
			FROM promised_amount
			WHERE status = 'mining' AND
					     del_block_id = 0 AND
						(cash_request_out_time = 0 OR cash_request_out_time > ?)
			GROUP BY currency_id`, "currency_id", "sum_amount", utils.Time()-c.Variables.Int64["cash_request_time"])

	// получаем кол-во майнеров по валютам
	promisedAmountMiners, err := c.GetMap(`
			SELECT currency_id, count(user_id) as count
			FROM (
					SELECT currency_id, user_id
					FROM promised_amount
					WHERE  del_block_id = 0 AND
								 del_mining_block_id = 0 AND
								 status IN ('mining', 'repaid')
					GROUP BY  user_id, currency_id
					) as t1
			GROUP BY  currency_id`, "currency_id", "count")

	// получаем кол-во анонимных юзеров по валютам
	walletsUsers, err := c.GetMap(`
			SELECT currency_id, count(user_id) as count
			FROM wallets
			WHERE amount > 0
			GROUP BY  currency_id`, "currency_id", "count")



	var userInfoWallets []utils.DCAmounts
	var promisedAmountListAccepted []utils.PromisedAmounts
	var credits map[string]string
	// поиск инфы о юзере
	userInfoId := int64(utils.StrToFloat64(c.Parameters["user_info_id"]))
	if userInfoId > 0 {
		userInfoWallets, err = c.GetBalances(userInfoId)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		// обещанные суммы юзера
		_, promisedAmountListAccepted, _, err = c.GetPromisedAmounts(userInfoId, c.Variables.Int64["cash_request_time"])
		// кредиты
		credits, err = c.GetMap(`
				SELECT sum(amount) as amount,
							 currency_id
				FROM credits
				WHERE from_user_id = ? AND
							 del_block_id = 0
				GROUP BY currency_id`, "amount", "currency_id", userInfoId)
	}

	/*
	 * Кол-во юзеров, сменивших ключ
	 * */
	countUsers, err := c.Single("SELECT count(user_id) FROM users WHERE log_id > 0").Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	/*
	 * %/год
	 * */
	currencyPct := make(map[int64]map[string]string)
	for currencyId, name := range c.CurrencyList {
		pct, err := c.OneRow("SELECT * FROM pct WHERE currency_id  =  ? ORDER BY block_id DESC", currencyId).Float64()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		currencyPct[currencyId] = make(map[string]string)
		currencyPct[currencyId]["name"] = name
		currencyPct[currencyId]["miner"] = utils.Float64ToStr(utils.Round((math.Pow(1+pct["miner"], 120)-1)*100, 6))
		currencyPct[currencyId]["user"] = utils.Float64ToStr(utils.Round((math.Pow(1+pct["user"], 120)-1)*100, 6))
		currencyPct[currencyId]["miner_year"] = utils.Float64ToStrPct(utils.Round((math.Pow(1+pct["miner"], 3600*24*365)-1)*100, 2))
		currencyPct[currencyId]["user_year"] = utils.Float64ToStrPct(utils.Round((math.Pow(1+pct["user"], 3600*24*365)-1)*100, 2))
	}

	/*
	 * Произошедшие сокращения
	 * */
	reduction, err := c.GetAll(`
			SELECT *
			FROM reduction
			ORDER BY time DESC
			LIMIT 20`, 20)
	for i := 0; i < len(reduction); i++ {
		if reduction[i]["type"] != "auto" {
			reduction[i]["type"] = "voting"
		}

		t := time.Unix(utils.StrToInt64(reduction[i]["time"]), 0)
		reduction[i]["time"] = t.Format(c.TimeFormat)
	}
	
	maxAmount := make( map[int64]float64 )
	for icur,_ := range sumWallets {
		max, err := c.Single("SELECT amount FROM max_promised_amounts WHERE currency_id  =  ? ORDER BY block_id DESC", icur ).Float64()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		maxAmount[icur] = max
		countMiners, err := c.Single("SELECT count(id) FROM promised_amount where currency_id = ? AND status='mining'", icur ).Int64()
		if countMiners < 1000 && maxAmount[icur] > float64(consts.MaxGreen[icur]) {
			maxAmount[icur] = float64(consts.MaxGreen[icur])
		}
	}

	TemplateStr, err := makeTemplate("statistic", "statistic", &StatisticPage{
		Lang:                       c.Lang,
		CurrencyList:               c.CurrencyListCf,
		UserInfoId:                 userInfoId,
		SumWallets:                 sumWallets,
		SumPromisedAmount:          sumPromisedAmount,
		PromisedAmountMiners:       promisedAmountMiners,
		WalletsUsers:               walletsUsers,
		UserInfoWallets:            userInfoWallets,
		Credits:                    credits,
		PromisedAmountListAccepted: promisedAmountListAccepted,
		MaxAmount:                  maxAmount,
		CountUsers:                 countUsers,
		CurrencyPct:                currencyPct,
		Reduction:                  reduction,
		UserId:                     c.SessUserId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
