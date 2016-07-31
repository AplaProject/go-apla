package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

type delAutoPaymentPage struct {
	SignData     string
	ShowSignData bool
	TxType       string
	TxTypeId     int64
	TimeNow      int64
	UserId       int64
	Alert        string
	Lang         map[string]string
	CountSignArr []int
	AutoId       int64
}

func (c *Controller) DelAutoPayment() (string, error) {

	txType := "DelAutoPayment"
	txTypeId := utils.TypeInt(txType)
	timeNow := time.Now().Unix()

	autoId := int64(utils.StrToFloat64(c.Parameters["auto_id"]))

	TemplateStr, err := makeTemplate("del_auto_payment", "delAutoPayment", &delAutoPaymentPage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		CountSignArr: c.CountSignArr,
		ShowSignData: c.ShowSignData,
		UserId:       c.SessUserId,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId,
		SignData:     "",
		AutoId:       autoId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
