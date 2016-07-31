package controllers

import (
	"errors"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"regexp"
	"strings"
)

func (c *Controller) EGatePm() (string, error) {

	c.r.ParseForm()

	fmt.Println(c.r.Form)

	sign := strings.ToUpper(string(utils.Md5(c.r.FormValue("PAYMENT_ID") + ":" + c.r.FormValue("PAYEE_ACCOUNT") + ":" + c.r.FormValue("PAYMENT_AMOUNT") + ":" + c.r.FormValue("PAYMENT_UNITS") + ":" + c.r.FormValue("PAYMENT_BATCH_NUM") + ":" + c.r.FormValue("PAYER_ACCOUNT") + ":" + strings.ToUpper(string(utils.Md5(c.EConfig["pm_s_key"]))) + ":" + c.r.FormValue("TIMESTAMPGMT"))))

	txTime := utils.StrToInt64(c.r.FormValue("TIMESTAMPGMT"))

	if sign != c.r.FormValue("V2_HASH") {
		return "", errors.New("Incorrect signature")
	}

	currencyId := int64(0)

	if c.r.FormValue("PAYMENT_UNITS") == "USD" {
		currencyId = 1001
	}
	if currencyId == 0 {
		return "", errors.New("Incorrect currencyId")
	}

	amount := utils.StrToFloat64(c.r.FormValue("PAYMENT_AMOUNT"))
	pmId := utils.StrToInt64(c.r.FormValue("PAYMENT_BATCH_NUM"))
	// проверим, не зачисляли ли мы уже это платеж
	existsId, err := c.Single(`SELECT id FROM e_adding_funds_pm WHERE id = ?`, pmId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if existsId != 0 {
		return "", errors.New("Incorrect PAYMENT_BATCH_NUM")
	}
	paymentInfo := c.r.FormValue("PAYMENT_ID")

	err = EPayment(paymentInfo, currencyId, txTime, amount, pmId, "pm", c.ECommission)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	return ``, nil
}

func EPayment(paymentInfo string, currencyId, txTime int64, amount float64, paymentId int64, paymentSystem string, eCommission float64) error {

	var userId int64
	r, _ := regexp.Compile(`(?i)token\-([0-9]+)`)
	t_ := r.FindStringSubmatch(paymentInfo)
	token := ""
	if len(t_) > 0 {
		token = t_[1]
	} else {
		userId = utils.StrToInt64(paymentInfo)
	}
	var buyCurrencyId int64
	if len(token) > 0 {
		err := utils.DB.ExecSql(`UPDATE e_tokens SET status = 'paid' WHERE id = ?`, token)
		if err != nil {
			return utils.ErrInfo(err)
		}
		data, err := utils.DB.OneRow(`SELECT user_id, buy_currency_id FROM e_tokens WHERE id = ?`, token).Int64()
		if err != nil {
			return utils.ErrInfo(err)
		}
		userId = data["user_id"]

		buyCurrencyId = data["buy_currency_id"]
	}
	err := utils.DB.ExecSql(`INSERT INTO e_adding_funds_`+paymentSystem+` (id, user_id, currency_id, time, amount) VALUES (?, ?, ?, ?, ?)`, paymentId, userId, currencyId, txTime, amount)
	if err != nil {
		return utils.ErrInfo(err)
	}
	if userId > 0 {
		err = utils.UpdEWallet(userId, currencyId, utils.Time(), amount, false)
		if err != nil {
			return utils.ErrInfo(err)
		}

		// автоматом создаем ордер, если это запрос через кошель Dcoin
		if len(token) > 0 {
			err = NewForexOrder(userId, amount, 1, currencyId, buyCurrencyId, "buy", eCommission)
			if err != nil {
				return utils.ErrInfo(err)
			}
		}
	}
	return nil
}
