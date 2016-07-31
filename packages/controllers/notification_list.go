package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"encoding/json"
	"fmt"
)

type notificationListPage struct {
	Lang            map[string]string
	LangInt         int64
	List            []map[string]string
}

var (
	nlCurrency map[int64]string = make(map[int64]string)
)

func Currency(currency int64) string {
	if ret,ok := nlCurrency[currency]; ok {
		return ret
	}
	ret,_ := utils.DB.Single(`SELECT name FROM currency where id=?`,currency ).String()
	if len(ret) > 0 {
		nlCurrency[currency] = ret
	}
	return ret
}


func (c *Controller) NotificationList() (string, error) {
	list, err := c.GetAll("SELECT * FROM notifications WHERE user_id = ? ORDER BY id DESC", 30, c.SessUserId )
//	list, err := c.GetAll("SELECT * FROM notifications ORDER BY id DESC", 150 )
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	
	for key,val := range list {
		var txt string
		switch utils.StrToInt64(val[`cmd_id`]) {
			case utils.ECMD_CASHREQ:
				var params utils.TypeNfyCashRequest
				json.Unmarshal([]byte(val[`params`]),&params)
				txt = fmt.Sprintf(`You"ve got the request for %.6f %s from user ID %d. It has to be repaid within the next 48 hours.`, 
				                params.Amount, Currency( params.CurrencyId ), params.FromUserId )
			case utils.ECMD_REFREADY:
				var params utils.TypeNfyRefReady
				json.Unmarshal([]byte(val[`params`]),&params)
				txt = fmt.Sprintf(`The referral key for the user ID %s is ready.`, params.RefId )
			case utils.ECMD_CHANGESTAT:
				var params utils.TypeNfyStatus
				json.Unmarshal([]byte(val[`params`]),&params)
				txt = fmt.Sprintf(`Your new status is %s.`, params.Status )
			case utils.ECMD_DCSENT:
				var params utils.TypeNfySent
				json.Unmarshal([]byte(val[`params`]),&params)
				txt = fmt.Sprintf(`You've sent %.6f d%s to ID %d.`, params.Amount, Currency( params.CurrencyId ), params.ToUserId )
			case utils.ECMD_DCCAME:
				var params utils.TypeNfyCame
				json.Unmarshal([]byte(val[`params`]),&params)
				if params.TypeTx != "node_commission" && params.Amount>0.000001 {
					txt = fmt.Sprintf(`You've got %.6f d%s from ID %d.`, params.Amount, Currency( params.CurrencyId ), params.FromUserId )
				}
		}
		list[key][`txt`] = txt
/*		for pk,pv := range params {
			list[key][pk] = pv
		}*/
	}
	c.ExecSql("UPDATE notifications SET isread=0 WHERE user_id = ?", c.SessUserId )
	
	TemplateStr, err := makeTemplate("notification_list", "notification_list", &notificationListPage{
		Lang:            c.Lang,
		LangInt:         c.LangInt,
		List:            list})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
