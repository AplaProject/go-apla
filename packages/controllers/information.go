package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type InformationPage struct {
	AlertMessages []map[string]string
	Lang          map[string]string
}

func (c *Controller) Information() (string, error) {

	var err error

	alertMessages_, err := c.GetAll(`SELECT * FROM alert_messages ORDER BY id DESC`, -1)
	var alertMessages []map[string]string
	for _, v := range alertMessages_ {
		show := false
		if v["currency_list"] != "ALL" {
			// проверим, есть ли у нас обещнные суммы с такой валютой
			amounts, err := c.Single("SELECT id FROM promised_amount WHERE currency_id IN (" + v["currency_list"] + ")").Int64()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			if amounts > 0 {
				show = true
			}
		} else {
			show = true
		}
		if show {
			alertMessages = append(alertMessages, map[string]string{"id": v["id"], "message": v["message"]})
		}
	}

	TemplateStr, err := makeTemplate("information", "information", &InformationPage{
		Lang:          c.Lang,
		AlertMessages: alertMessages})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
