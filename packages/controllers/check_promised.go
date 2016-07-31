// check_promised
package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"encoding/json"
)

func (c *Controller) CheckPromised() (string, error) {

	resval := false
	result := func(msg, data string, success bool ) (string, error) {
		res, err := json.Marshal( answerJson{Result:resval, Error: msg,
		                          Data: data, Success: success})
		return string(res), err
	}
	
	userId := utils.StrToInt64( c.r.FormValue(`user_id`))
	amount := utils.StrToFloat64( c.r.FormValue(`amount`))
	currencyId := utils.StrToInt64( c.r.FormValue(`currency_id`))
	
	_, promised, _, err := c.GetPromisedAmounts(userId, c.Variables.Int64["cash_request_time"])
	if err != nil {
		return result(err.Error(), ``, false)
	}
	var promisedAmount float64
	for _, item := range promised {
		if item.CurrencyId == currencyId && item.Status == `mining` {
			promisedAmount = item.Amount
			countMiners, err := c.Single("SELECT count(id) FROM promised_amount where currency_id = ? AND status='mining'", currencyId).Int64()
			if err != nil {
				return result(err.Error(), ``, false)
			}
			if countMiners < 1000 && promisedAmount > float64(consts.MaxGreen[currencyId]) {
				promisedAmount = float64(consts.MaxGreen[currencyId])
			}
			if promisedAmount >= amount {
				resval = true
				return result(``, ``, true)
			}
		}
	}
	return result( ``, utils.IntToStr( int( promisedAmount )), true )
}
