package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type anonymMoneyTransferPage struct {
	Lang                  map[string]string
	Title                 string
	CountSign             int
	CountSignArr          []int
	SignData              string
	ShowSignData          bool
	TxType       string
	TxTypeId     int64
	TimeNow      int64
	WalletId int64
}

func (c *Controller) AnonymMoneyTransfer() (string, error) {

	txType := "DLTTransfer"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	TemplateStr, err := makeTemplate("anonym_money_transfer", "anonymMoneyTransfer", &anonymMoneyTransferPage{
		CountSignArr:          c.CountSignArr,
		CountSign:             c.CountSign,
		Lang:                  c.Lang,
		Title:                 "anonymMoneyTransfer",
		ShowSignData:          c.ShowSignData,
		SignData:              "",
		WalletId: c.SessWalletId,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
