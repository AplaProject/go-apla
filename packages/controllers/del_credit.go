package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

type delCreditPage struct {
	SignData     string
	ShowSignData bool
	TxType       string
	TxTypeId     int64
	TimeNow      int64
	UserId       int64
	Alert        string
	Lang         map[string]string
	CountSignArr []int
	CreditId     int64
}

func (c *Controller) DelCredit() (string, error) {

	txType := "DelCredit"
	txTypeId := utils.TypeInt(txType)
	timeNow := time.Now().Unix()

	creditId := int64(utils.StrToFloat64(c.Parameters["credit_id"]))

	TemplateStr, err := makeTemplate("del_credit", "delCredit", &delCreditPage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		CountSignArr: c.CountSignArr,
		ShowSignData: c.ShowSignData,
		UserId:       c.SessUserId,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId,
		SignData:     "",
		CreditId:     creditId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
