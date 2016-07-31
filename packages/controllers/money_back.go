package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type moneyBackPage struct {
	Alert        string
	SignData     string
	ShowSignData bool
	CountSignArr []int
	Lang         map[string]string
	UserId       int64
	OrderId      int64
	Amount       float64
	Arbitrator   int64
	Li           string
	Redirect     string
	TxType       string
	TxTypeId     int64
	TimeNow      int64
}

func (c *Controller) MoneyBack() (string, error) {

	txType := "MoneyBack"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	orderId := int64(utils.StrToFloat64(c.Parameters["order_id"]))
	amount := utils.StrToFloat64(c.Parameters["amount"])
	arbitrator := int64(utils.StrToFloat64(c.Parameters["arbitrator"]))
	var li, redirect string
	if arbitrator > 0 {
		li = `<li><a href="#arbitrationArbitrator">` + c.Lang["i_arbitrator"] + `</a></li>`
		redirect = `arbitrationArbitrator`
	} else {
		li = `<li><a href="#arbitrationArbitrator">` + c.Lang["i_seller"] + `</a></li>`
		redirect = `arbitrationSeller`
	}

	TemplateStr, err := makeTemplate("money_back", "moneyBack", &moneyBackPage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		ShowSignData: c.ShowSignData,
		SignData:     "",
		UserId:       c.SessUserId,
		OrderId:      orderId,
		Amount:       amount,
		Arbitrator:   arbitrator,
		Li:           li,
		Redirect:     redirect,
		CountSignArr: c.CountSignArr,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
