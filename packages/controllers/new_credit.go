package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

type newCreditPage struct {
	SignData     string
	ShowSignData bool
	TxType       string
	TxTypeId     int64
	TimeNow      int64
	UserId       int64
	Alert        string
	Lang         map[string]string
	CountSignArr []int
	CreditId     float64
	CurrencyList map[int64]string
}

func (c *Controller) NewCredit() (string, error) {

	txType := "NewCredit"
	txTypeId := utils.TypeInt(txType)
	timeNow := time.Now().Unix()

	creditId := utils.Round(utils.StrToFloat64(c.Parameters["credit_id"]), 0)

	TemplateStr, err := makeTemplate("new_credit", "newCredit", &newCreditPage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		CountSignArr: c.CountSignArr,
		ShowSignData: c.ShowSignData,
		UserId:       c.SessUserId,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId,
		SignData:     "",
		CreditId:     creditId,
		CurrencyList: c.CurrencyList})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
