package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type changeMoneyBackTimePage struct {
	Alert        string
	SignData     string
	ShowSignData bool
	CountSignArr []int
	Lang         map[string]string
	UserId       int64
	OrderId      int64
	Days         int64
	TxType       string
	TxTypeId     int64
	TimeNow      int64
}

func (c *Controller) ChangeMoneyBackTime() (string, error) {

	txType := "ChangeMoneyBackTime"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	orderId := int64(utils.StrToFloat64(c.Parameters["order_id"]))
	days := int64(utils.StrToFloat64(c.Parameters["days"]))

	TemplateStr, err := makeTemplate("change_money_back_time", "changeMoneyBackTime", &changeMoneyBackTimePage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		ShowSignData: c.ShowSignData,
		SignData:     "",
		UserId:       c.SessUserId,
		OrderId:      orderId,
		Days:         days,
		CountSignArr: c.CountSignArr,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
