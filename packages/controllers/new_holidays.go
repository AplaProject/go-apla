package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type newHolidaysPage struct {
	SignData     string
	ShowSignData bool
	TxType       string
	TxTypeId     int64
	TimeNow      int64
	UserId       int64
	Alert        string
	Lang         map[string]string
	CountSignArr []int
}

func (c *Controller) NewHolidays() (string, error) {

	var err error

	txType := "NewHolidays"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	TemplateStr, err := makeTemplate("new_holidays", "newHolidays", &newHolidaysPage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		TxTypeId:     txTypeId,
		TimeNow:      timeNow,
		CountSignArr: c.CountSignArr,
		ShowSignData: c.ShowSignData,
		SignData:     ""})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
