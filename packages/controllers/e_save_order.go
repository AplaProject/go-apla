package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"strings"
)

func (c *Controller) ESaveOrder() (string, error) {

	if c.SessUserId == 0 {
		return "", errors.New(c.Lang["sign_up_please"])
	}
	c.r.ParseForm()
	sellCurrencyId := utils.StrToInt64(c.r.FormValue("sell_currency_id"))
	buyCurrencyId := utils.StrToInt64(c.r.FormValue("buy_currency_id"))
	amount := utils.StrToFloat64(c.r.FormValue("amount"))
	sellRate := utils.StrToFloat64(c.r.FormValue("sell_rate"))
	orderType := c.r.FormValue("type")
	// можно ли торговать такими валютами
	checkCurrency, err := c.Single("SELECT count(id) FROM e_currency WHERE id IN (?, ?)", sellCurrencyId, buyCurrencyId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if checkCurrency != 2 {
		return "", errors.New("Currency error")
	}
	if orderType != "sell" && orderType != "buy" {
		return "", errors.New("Type error")
	}
	if amount == 0 {
		return "", errors.New(c.Lang["amount_error"])
	}
	if amount < 0.001 && sellCurrencyId < 1000 {
		return "", errors.New(strings.Replace(c.Lang["save_order_min_amount"], "[amount]", "0.001", -1))
	}
	if sellRate < 0.0001 {
		return "", errors.New(strings.Replace(c.Lang["save_order_min_price"], "[price]", "0.0001", -1))
	}
	reductionLock, err := utils.EGetReductionLock()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if reductionLock > 0 {
		return "", errors.New(strings.Replace(c.Lang["creating_orders_unavailable"], "[minutes]", "30", -1))
	}

	// нужно проверить, есть ли нужная сумма на счету юзера
	userAmountAndProfit := utils.EUserAmountAndProfit(c.SessUserId, sellCurrencyId)
	if userAmountAndProfit < amount {
		return "", errors.New(c.Lang["not_enough_money"] + " (" + utils.Float64ToStr(userAmountAndProfit) + "<" + utils.Float64ToStr(amount) + ")" + strings.Replace(c.Lang["add_funds_link"], "[currency]", "USD", -1))
	}

	err = NewForexOrder(c.SessUserId, amount, sellRate, sellCurrencyId, buyCurrencyId, orderType, utils.StrToMoney(c.EConfig["commission"]))
	if err != nil {
		return "", utils.ErrInfo(err)
	} else {
		return utils.JsonAnswer(c.Lang["order_created"], "success").String(), nil
	}

	return ``, nil
}
