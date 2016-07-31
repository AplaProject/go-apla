package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"

	"fmt"
	"math"
)

type miningPromisedAmountPage struct {
	Alert            string
	SignData         string
	ShowSignData     bool
	TxType           string
	TxTypeId         int64
	TimeNow          int64
	UserId           int64
	Lang             map[string]string
	CountSignArr     []int
	LastTxFormatted  string
	CurrencyList     map[int64]string
	DelId            string
	Amount           float64
	PromisedAmountId int64
}

func (c *Controller) MiningPromisedAmount() (string, error) {

	txType := "Mining"
	txTypeId := utils.TypeInt(txType)
	timeNow := time.Now().Unix()
	amount := utils.StrToMoney(c.Parameters["amount"])
	amount = math.Floor(amount*100) / 100
	promisedAmountId := int64(utils.StrToFloat64(c.Parameters["promised_amount_id"]))
	log.Debug("c.Parameters[promised_amount_id]):", c.Parameters["promised_amount_id"])
	log.Debug("promisedAmountId:", promisedAmountId)
	TemplateStr, err := makeTemplate("mining_promised_amount", "miningPromisedAmount", &miningPromisedAmountPage{
		Alert:            c.Alert,
		Lang:             c.Lang,
		CountSignArr:     c.CountSignArr,
		ShowSignData:     c.ShowSignData,
		UserId:           c.SessUserId,
		TimeNow:          timeNow,
		TxType:           txType,
		TxTypeId:         txTypeId,
		SignData:         fmt.Sprintf("%v,%v,%v,%v,%v", txTypeId, timeNow, c.SessUserId, promisedAmountId, amount),
		Amount:           amount,
		PromisedAmountId: promisedAmountId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
