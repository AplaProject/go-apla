package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"strings"
)

type holidaysListPage struct {
	SignData           string
	ShowSignData       bool
	Alert              string
	Lang               map[string]string
	CountSignArr       []int
//	LastTxFormatted    string
	LimitsText         string
	MyHolidaysPending  []map[string]string
	MyHolidaysAccepted []map[string]string
}

func (c *Controller) HolidaysList() (string, error) {

	var err error

	var myHolidaysPending []map[string]string
	if c.SessRestricted == 0 {
		// те, что еще не попали в Dc-сеть
		myHolidaysPending, err = c.GetAll(`SELECT * FROM `+c.MyPrefix+`my_holidays ORDER BY id DESC`, -1)
	}

	myHolidaysAccepted, err := c.GetAll(`SELECT * FROM holidays WHERE user_id = ?`, -1, c.SessUserId)

	limitsText := strings.Replace(c.Lang["limits_text"], "[limit]", utils.Int64ToStr(c.Variables.Int64["limit_holidays"]), -1)
	limitsText = strings.Replace(limitsText, "[period]", c.Periods[c.Variables.Int64["limit_holidays_period"]], -1)

/*	last_tx, err := c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"NewHolidays"}), 3, c.TimeFormat)
	lastTxFormatted := ""
	if len(last_tx) > 0 {
		lastTxFormatted, _ = utils.MakeLastTx(last_tx, c.Lang)
	}*/

	TemplateStr, err := makeTemplate("holidays_list", "holidaysList", &holidaysListPage{
		Alert:              c.Alert,
		Lang:               c.Lang,
		CountSignArr:       c.CountSignArr,
		ShowSignData:       c.ShowSignData,
		SignData:           "",
//		LastTxFormatted:    lastTxFormatted,
		LimitsText:         limitsText,
		MyHolidaysPending:  myHolidaysPending,
		MyHolidaysAccepted: myHolidaysAccepted})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
