package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"sort"
	"strings"
	"time"
)

type eMainPage struct {
	AlertMessages    []map[string]string
	Lang             map[string]string
	CurrencyList     map[int64]string
	Commission       string
	Members          int64
	SellMax          float64
	BuyMin           float64
	EOrdersSell      []map[string]float64
	EOrdersBuy       []map[string]float64
	UserId           int64
	DcCurrency       string
	Currency         string
	DcCurrencyId     int64
	CurrencyId       int64
	TradeHistory     []map[string]string
	CurrencyListPair map[int64][]int64
	CommissionText   string
}

func (c *Controller) EMain() (string, error) {

	var err error

	dcCurrencyId := utils.StrToInt64(c.Parameters["dc_currency_id"])
	currencyId := utils.StrToInt64(c.Parameters["currency_id"])
	if dcCurrencyId == 0 {
		dcCurrencyId = 72
	}
	if currencyId == 0 {
		currencyId = 1001
	}

	// все валюты, с которыми работаем
	currencyList, err := utils.EGetCurrencyList()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	log.Debug("%v", currencyList)

	// работаем только с теми валютами, которые есть у нас в списке
	if len(currencyList[dcCurrencyId]) == 0 || len(currencyList[currencyId]) == 0 {
		return "", utils.ErrInfo("incorrect currency")
	}

	// пары валют для меню
	currencyListPair, err := eGetCurrencyPair()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	dcCurrency := currencyList[dcCurrencyId]
	currency := currencyList[currencyId]

	// история сделок
	var tradeHistory []map[string]string

	rows, err := c.Query(c.FormatQuery(`
			SELECT sell_currency_id, sell_rate, amount, time
			FROM e_trade
			WHERE ((sell_currency_id = ? AND buy_currency_id = ?) OR (sell_currency_id = ? AND buy_currency_id = ?)) AND main = 1
			ORDER BY time DESC
			LIMIT 40
			`), dcCurrencyId, currencyId, currencyId, dcCurrencyId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var sellCurrencyId, eTime int64
		var sellRate, amount float64
		err = rows.Scan(&sellCurrencyId, &sellRate, &amount, &eTime)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		var eType string
		var eAmount float64
		var eTotal float64
		if sellCurrencyId == dcCurrencyId {
			eType = "sell"
			sellRate = 1 / sellRate
			eAmount = amount
			eTotal = amount * sellRate
		} else {
			eType = "buy"
			eAmount = amount * (1 / sellRate)
			eTotal = amount
		}
		t := time.Unix(eTime, 0)
		tradeHistory = append(tradeHistory, map[string]string{"Time": t.Format(c.TimeFormat), "Type": eType, "SellRate": utils.ClearNull(utils.Float64ToStr(sellRate), 4), "Amount": utils.ClearNull(utils.Float64ToStr(eAmount), 4), "Total": utils.ClearNull(utils.Float64ToStr(eTotal), 4)})
	}

	// активные ордеры на продажу
	var orders eOrders
	rows, err = c.Query(c.FormatQuery(`
			SELECT sell_rate, amount
			FROM e_orders
			WHERE (sell_currency_id = ? AND buy_currency_id = ?) AND
						empty_time = 0 AND
						del_time = 0 AND
						amount > 0
			ORDER BY sell_rate DESC
			LIMIT 100
			`), dcCurrencyId, currencyId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	defer rows.Close()
	// мин. цена покупки
	var buyMin float64
	for rows.Next() {
		var sellRate, amount float64
		err = rows.Scan(&sellRate, &amount)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if orders.Sell == nil {
			orders.Sell = make(map[float64]float64)
		}
		sellRate = utils.ClearNullFloat64(1/sellRate, 6)
		orders.Sell[sellRate] = utils.ClearNullFloat64(orders.Sell[sellRate]+amount, 6)
		if buyMin == 0 {
			buyMin = sellRate
		} else if sellRate < buyMin {
			buyMin = sellRate
		}
	}

	keys := []float64{}
	for k := range orders.Sell {
		keys = append(keys, k)
	}
	sort.Float64s(keys)
	var eOrdersSell []map[string]float64
	for _, k := range keys {
		eOrdersSell = append(eOrdersSell, map[string]float64{"sell_rate": k, "amount": orders.Sell[k]})
	}

	// активные ордеры на покупку
	rows, err = c.Query(c.FormatQuery(`
			SELECT sell_rate, amount
			FROM e_orders
			WHERE (sell_currency_id = ? AND buy_currency_id = ?) AND
					empty_time = 0 AND
					del_time = 0 AND
					amount > 0
			ORDER BY sell_rate ASC
			LIMIT 100
			`), currencyId, dcCurrencyId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	defer rows.Close()
	// мин. цена продажи
	var sellMax float64
	for rows.Next() {
		var sellRate, amount float64
		err = rows.Scan(&sellRate, &amount)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if orders.Buy == nil {
			orders.Buy = make(map[float64]float64)
		}
		sellRate = utils.ClearNullFloat64(sellRate, 6)
		orders.Buy[sellRate] = utils.ClearNullFloat64(orders.Buy[sellRate]+amount*(1/sellRate), 6)
		if sellMax == 0 {
			sellMax = sellRate
		} else if sellRate > sellMax {
			sellMax = sellRate
		}
	}
	var keysR sort.Float64Slice
	for k := range orders.Buy {
		keysR = append(keysR, k)
	}
	sort.Sort(sort.Reverse(keysR))
	var eOrdersBuy []map[string]float64
	for _, k := range keysR {
		eOrdersBuy = append(eOrdersBuy, map[string]float64{"sell_rate": k, "amount": orders.Buy[k]})
	}

	// комиссия
	commission := c.EConfig["commission"]
	commissionText := strings.Replace(c.Lang["commission_text"], "[commission]", commission, -1)

	// кол-во юзеров
	members, err := c.Single(`SELECT count(*) FROM e_users`).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	TemplateStr, err := makeTemplate("e_main", "eMain", &eMainPage{
		Lang:             c.Lang,
		Commission:       commission,
		Members:          members,
		SellMax:          sellMax,
		BuyMin:           buyMin,
		EOrdersSell:      eOrdersSell,
		EOrdersBuy:       eOrdersBuy,
		DcCurrency:       dcCurrency,
		Currency:         currency,
		DcCurrencyId:     dcCurrencyId,
		UserId:           c.SessUserId,
		TradeHistory:     tradeHistory,
		CurrencyId:       currencyId,
		CommissionText:   commissionText,
		CurrencyListPair: currencyListPair,
		CurrencyList:     currencyList})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}

type eOrders struct {
	Sell map[float64]float64
	Buy  map[float64]float64
}

func eGetCurrencyPair() (map[int64][]int64, error) {
	currencyList := make(map[int64][]int64)
	rows, err := utils.DB.Query(`
			SELECT id,
				   currency,
				   dc_currency
			FROM e_currency_pair
			ORDER BY id
			`)
	if err != nil {
		return currencyList, utils.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id, currency, dc_currency int64
		err = rows.Scan(&id, &currency, &dc_currency)
		if err != nil {
			return currencyList, utils.ErrInfo(err)
		}
		currencyList[id] = []int64{currency, dc_currency}
	}
	return currencyList, nil
}
