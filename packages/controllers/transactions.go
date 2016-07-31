// transactions
package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"time"
)

type transactionsPage struct {
	UserId            int64
	Lang              map[string]string
	List              []map[string]string
}

func (c *Controller) Transactions() (string, error) {

	var err error

	list, err := c.GetLastTx(c.SessUserId, nil, 20, c.TimeFormat)
	for ind, data := range list {
		var result string
		if utils.StrToInt64(data["block_id"]) > 0 {
			result = c.Lang["in_the_block"] + " " + data["block_id"]
		} else if len(data["error"]) > 0 {
			result = "Error: " + data["error"] 
		} else if (len(data["queue_tx"]) == 0 && len(data["tx"]) == 0) || time.Now().Unix()-utils.StrToInt64(data["time_int"]) > 7200 {
			result += c.Lang["lost"]
		} else {
			result += c.Lang["status_pending"]
		}
		list[ind][`result`] = result
		list[ind][`txtype`] = consts.TxTypes[utils.StrToInt(data[`type`])]
	}

	TemplateStr, err := makeTemplate("transactions", "transactions", &transactionsPage{
		Lang:              c.Lang,
		List:              list,
		UserId:            c.SessUserId,
	})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
