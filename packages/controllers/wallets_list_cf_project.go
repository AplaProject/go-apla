package controllers

import (
	"encoding/json"
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"math"
	"time"
)

func (c *Controller) WalletsListCfProject() (string, error) {

	var err error
	c.r.ParseForm()

	projectId := utils.StrToInt64(c.r.FormValue("project_id"))
	if projectId == 0 {
		return "", errors.New("projectId == 0")
	}
	cfProject, err := c.OneRow("SELECT id, amount, currency_id, end_time FROM cf_projects WHERE del_block_id = 0 AND id  =  ?", projectId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	//id, endTime, langId int64, amount float64, levelUp string
	CfProjectData, err := c.GetCfProjectData(projectId, utils.StrToInt64(cfProject["end_time"]), c.LangInt, utils.StrToFloat64(cfProject["amount"]), "")
	for k, v := range CfProjectData {
		cfProject[k] = v
	}

	// сколько у нас есть DC данной валюты
	wallet, err := c.OneRow("SELECT amount, currency_id,  last_update FROM wallets WHERE user_id  =  ? AND currency_id  =  ?", c.SessUserId, cfProject["currency_id"]).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if len(wallet) > 0 {
		amount := utils.StrToMoney(wallet["amount"])
		profit, err := c.CalcProfitGen(utils.StrToInt64(wallet["currency_id"]), amount, c.SessUserId, utils.StrToInt64(wallet["last_update"]), time.Now().Unix(), "wallet")
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		amount += profit
		amount = math.Floor(utils.Round(amount, 3)*100) / 100
		forexOrdersAmount, err := c.Single("SELECT sum(amount) FROM forex_orders WHERE user_id  =  ? AND sell_currency_id  =  ? AND del_block_id  =  0", c.SessUserId, wallet["currency_id"]).Float64()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		amount -= forexOrdersAmount
		cfProject["wallet_amount"] = utils.Float64ToStrPct(amount)
	} else {
		cfProject["wallet_amount"] = "0"
	}

	cfProject["currency"] = c.CurrencyList[utils.StrToInt64(cfProject["currency_id"])]

	newmap := make(map[string]interface{})
	for k, v := range cfProject {
		newmap[k] = v
	}

	// наличие описаний
	newmap["lang"], err = c.GetMap("SELECT id,  lang_id FROM cf_projects_data	WHERE project_id = ?", "id", "lang_id", projectId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	result, err := json.Marshal(newmap)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return string(result), nil
}
