package controllers

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

type delPromisedAmountPage struct {
	Alert           string
	SignData        string
	ShowSignData    bool
	TxType          string
	TxTypeId        int64
	TimeNow         int64
	UserId          int64
	Lang            map[string]string
	CountSignArr    []int
	LastTxFormatted string
	CurrencyList    map[int64]string
	DelId           string
}

func (c *Controller) DelPromisedAmount() (string, error) {

	txType := "DelPromisedAmount"
	txTypeId := utils.TypeInt(txType)
	timeNow := time.Now().Unix()
	delId := c.Parameters["del_id"]

	TemplateStr, err := makeTemplate("del_promised_amount", "delPromisedAmount", &delPromisedAmountPage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		CountSignArr: c.CountSignArr,
		ShowSignData: c.ShowSignData,
		UserId:       c.SessUserId,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId,
		SignData:     fmt.Sprintf("%d,%d,%d,%s", txTypeId, timeNow, c.SessUserId, delId),
		DelId:        delId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
