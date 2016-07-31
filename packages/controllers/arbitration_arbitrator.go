package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

type arbitrationArbitratorPage struct {
	Alert           string
	SignData        string
	ShowSignData    bool
	TxType          string
	TxTypeId        int64
	TimeNow         int64
	UserId          int64
	Lang            map[string]string
	CountSignArr    []int
//	LastTxFormatted string
	CurrencyList    map[int64]string
	MinerId         int64
	MyOrders        []map[string]string
}

func (c *Controller) ArbitrationArbitrator() (string, error) {

	log.Debug("ArbitrationArbitrator")

	txType := "MoneyBackRequest"
	txTypeId := utils.TypeInt(txType)
	timeNow := time.Now().Unix()

	myOrders, err := c.GetAll(`
			SELECT *
			FROM orders
			WHERE (arbitrator0 = ? OR arbitrator1 = ? OR arbitrator2 = ? OR arbitrator3 = ? OR arbitrator4 = ?) AND
						 status = 'refund'
			ORDER BY time DESC
			LIMIT 20
	`, 20, c.SessUserId, c.SessUserId, c.SessUserId, c.SessUserId, c.SessUserId)
	for k, data := range myOrders {
		if c.SessRestricted == 0 {
			data_, err := c.OneRow(`
						SELECT comment,
									 comment_status
						FROM `+c.MyPrefix+`my_comments
						WHERE id = ? AND
									 type = 'arbitrator'
						LIMIT 1
				`, data["id"]).String()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			data["comment"] = data_["comment"]
			data["comment_status"] = data_["comment_status"]
		} else {
			data["comment"] = ""
			data["comment_status"] = "decrypted"
		}
		myOrders[k] = data
	}

/*	last_tx, err := c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"ChangeArbitratorConditions", "MoneyBack"}), 3, c.TimeFormat)
	lastTxFormatted := ""
	if len(last_tx) > 0 {
		lastTxFormatted, _ = utils.MakeLastTx(last_tx, c.Lang)
	}*/

	TemplateStr, err := makeTemplate("arbitration_arbitrator", "arbitrationArbitrator", &arbitrationArbitratorPage{
		Alert:           c.Alert,
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
		MinerId:         c.MinerId,
		MyOrders:        myOrders})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
