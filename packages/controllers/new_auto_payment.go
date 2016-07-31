package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

type newAutoPaymentPage struct {
	SignData     string
	ShowSignData bool
	TxType       string
	TxTypeId     int64
	TimeNow      int64
	UserId       int64
	Alert        string
	Lang         map[string]string
	CountSignArr []int
	CurrencyList map[int64]string
}

func (c *Controller) NewAutoPayment() (string, error) {

	txType := "NewAutoPayment"
	txTypeId := utils.TypeInt(txType)
	timeNow := time.Now().Unix()

	TemplateStr, err := makeTemplate("new_auto_payment", "newAutoPayment", &newAutoPaymentPage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		CountSignArr: c.CountSignArr,
		ShowSignData: c.ShowSignData,
		UserId:       c.SessUserId,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId,
		SignData:     "",
		CurrencyList: c.CurrencyList})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
