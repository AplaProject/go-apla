package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type moneyBackRequestPage struct {
	Alert        string
	SignData     string
	ShowSignData bool
	CountSignArr []int
	Lang         map[string]string
	UserId       int64
	OrderId      int64
	Order        map[string]string
	TxType       string
	TxTypeId     int64
	TimeNow      int64
}

func (c *Controller) MoneyBackRequest() (string, error) {

	txType := "MoneyBackRequest"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	orderId := int64(utils.StrToFloat64(c.Parameters["order_id"]))
	order, err := c.OneRow("SELECT * FROM orders WHERE id  =  ?", orderId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	TemplateStr, err := makeTemplate("money_back_request", "moneyBackRequest", &moneyBackRequestPage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		ShowSignData: c.ShowSignData,
		SignData:     "",
		UserId:       c.SessUserId,
		OrderId:      orderId,
		Order:        order,
		CountSignArr: c.CountSignArr,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
