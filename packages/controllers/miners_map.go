package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"errors"
	"fmt"
)

func (c *Controller) MinersMap() (string, error) {

	if c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	c.r.ParseForm()

	log.Debug("min_amount:", c.r.FormValue("min_amount"))
	minAmount := utils.StrToFloat64(c.r.FormValue("min_amount"))
	log.Debug("minAmount:", minAmount)
	if !utils.CheckInputData(c.r.FormValue("min_amount"), "amount") {
		return "", errors.New("Incorrect min_amount")
	}
	currencyId := utils.StrToInt64(c.r.FormValue("currency_id"))
	if !utils.CheckInputData(c.r.FormValue("currency_id"), "currency_id") {
		return "", errors.New("Incorrect currency_id")
	}
	paymentSystemId := utils.StrToInt64(c.r.FormValue("payment_system_id"))
	if !utils.CheckInputData(c.r.FormValue("payment_system_id"), "int") {
		return "", errors.New("Incorrect payment_system_id")
	}

//	maxPromisedAmounts, err := c.Single("SELECT amount FROM max_promised_amounts WHERE currency_id  =  ? ORDER BY time DESC", currencyId).Float64()
	maxPromisedAmounts, err := c.GetMaxPromisedAmount(currencyId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	// пока нет хотя бы 1000 майнеров по этой валюте, ограничиваем размер обещанной суммы
	countMiners, err := c.Single("SELECT count(id) FROM promised_amount where currency_id = ? AND status='mining'", currencyId).Int64()
	if countMiners < 1000 {
		maxPromisedAmounts = float64(consts.MaxGreen[currencyId])
	}

	addSql := ""
	if paymentSystemId > 0 {
		addSql = fmt.Sprintf(` (ps1 = %d OR ps2 = %d OR ps3 = %d OR ps4 = %d OR ps5 = %d) AND`, paymentSystemId, paymentSystemId, paymentSystemId, paymentSystemId, paymentSystemId)
	}

	rows, err := c.Query(c.FormatQuery(`
			SELECT  amount,
						 latitude,
						 longitude,
						 promised_amount.user_id
			FROM promised_amount
			LEFT JOIN miners_data ON miners_data.user_id = promised_amount.user_id
			WHERE  promised_amount.status = 'mining' AND
						 currency_id = ? AND
						  `+addSql+`
						  promised_amount.user_id != ? AND
						  del_block_id = 0
	`), currencyId, c.SessUserId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	result := ""
	defer rows.Close()
	for rows.Next() {
		var amount float64
		var latitude, longitude string
		var user_id int64
		err = rows.Scan(&amount, &latitude, &longitude, &user_id)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		repaid, err := c.Single("SELECT amount FROM promised_amount WHERE status  =  'repaid' AND currency_id  =  ? AND user_id  =  ? AND del_block_id  =  0", currencyId, user_id).Float64()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		var returnAmount float64
		if repaid+amount < maxPromisedAmounts {
			returnAmount = amount
		} else {
			returnAmount = maxPromisedAmounts - repaid
		}
		if returnAmount <= 0 {
			continue
		}
		result += fmt.Sprintf(`{"user_id": %v, "amount": %v, "longitude": %v, "latitude": %v},`, user_id, returnAmount, longitude, latitude)
	}
	if len(result) > 0 {
		result = result[:len(result)-1]
	}
	log.Debug(result)
	result = `{ "info": [` + result + `]}`
	return result, nil
}
