package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) EDelOrder() (string, error) {

	c.r.ParseForm()
	orderId := utils.StrToInt64(c.r.FormValue("order_id"))

	// возвращаем сумму ордера на кошелек + возращаем комиссию.
	order, err := utils.DB.OneRow("SELECT amount, sell_currency_id FROM e_orders WHERE id  =  ? AND user_id  =  ? AND del_time  =  0 AND empty_time  =  0", orderId, c.SessUserId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if len(order) == 0 {
		return "", utils.ErrInfo("order_id error")
	}
	sellCurrencyId := utils.StrToInt64(order["sell_currency_id"])
	amount := utils.StrToFloat64(order["amount"])

	amountAndCommission := utils.StrToFloat64(order["amount"]) / (1 - c.ECommission/100)
	// косиссия биржи
	commission := amountAndCommission - amount
	err = userLock(c.SessUserId)
	if err != nil {
		return "", err
	}

	// отмечаем, что ордер удален
	err = utils.DB.ExecSql("UPDATE e_orders SET del_time = ? WHERE id = ? AND user_id = ?", utils.Time(), orderId, c.SessUserId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	// возвращаем остаток ордера на кошель
	userAmount := utils.EUserAmountAndProfit(c.SessUserId, sellCurrencyId)
	err = utils.DB.ExecSql("UPDATE e_wallets SET amount = ?, last_update = ? WHERE user_id = ? AND currency_id = ?", userAmount+amountAndCommission, utils.Time(), c.SessUserId, sellCurrencyId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	// вычитаем комиссию биржи
	userAmount = utils.EUserAmountAndProfit(1, sellCurrencyId)
	err = utils.DB.ExecSql("UPDATE e_wallets SET amount = ?, last_update = ? WHERE user_id = 1 AND currency_id = ?", userAmount-commission, utils.Time(), sellCurrencyId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	userUnlock(c.SessUserId)

	return ``, nil
}
