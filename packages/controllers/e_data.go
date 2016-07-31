package controllers

import (
	"encoding/json"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) EData() (string, error) {

	c.w.Header().Set("Access-Control-Allow-Origin", "*")

	// сколько всего продается DC
	eOrders, err := c.GetAll(`SELECT sell_currency_id, sum(amount) as amount FROM e_orders  WHERE sell_currency_id < 1000 AND empty_time = 0 GROUP BY sell_currency_id`, 100)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	values := ""
	for _, data := range eOrders {
		values += utils.ClearNull(data["amount"], 0) + ` d` + c.CurrencyList[utils.StrToInt64(data["sell_currency_id"])] + `, `
	}
	if len(values) > 0 {
		values = values[:len(values)-2]
	}
	ps, err := c.Single(`SELECT value FROM e_config WHERE name = 'ps'`).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	jsonData, err := json.Marshal(map[string]string{"values": values, "ps": ps})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return string(jsonData), nil

}
