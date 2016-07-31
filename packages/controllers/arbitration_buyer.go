package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

type arbitrationBuyerPage struct {
	SignData        string
	ShowSignData    bool
	TxType          string
	TxTypeId        int64
	TimeNow         int64
	UserId          int64
	Alert           string
	Lang            map[string]string
	CountSignArr    []int
//	LastTxFormatted string
	CurrencyList    map[int64]string
	MyOrders        []map[string]string
}

func (c *Controller) ArbitrationBuyer() (string, error) {

	log.Debug("ArbitrationBuyer")

	txType := "money_back_request"
	txTypeId := utils.TypeInt(txType)
	timeNow := time.Now().Unix()

	myOrders, err := c.GetAll(`
			SELECT *
			FROM orders
			WHERE buyer = ?
			ORDER BY time DESC
			LIMIT 20
			`, 20, c.SessUserId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

/*	last_tx, err := c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"ChangeSellerHoldBack"}), 3, c.TimeFormat)
	lastTxFormatted := ""
	if len(last_tx) > 0 {
		lastTxFormatted, _ = utils.MakeLastTx(last_tx, c.Lang)
	}*/

	TemplateStr, err := makeTemplate("arbitration_buyer", "arbitrationBuyer", &arbitrationBuyerPage{
		Lang:            c.Lang,
		CountSignArr:    c.CountSignArr,
		ShowSignData:    c.ShowSignData,
		UserId:          c.SessUserId,
		TimeNow:         timeNow,
		TxType:          txType,
		TxTypeId:        txTypeId,
		SignData:        "",
//		LastTxFormatted: lastTxFormatted,
		CurrencyList:    c.CurrencyList,
		MyOrders:        myOrders})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
