package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

type AutoPaymentsPage struct {
	SignData        string
	ShowSignData    bool
	TxType          string
	TxTypeId        int64
	TimeNow         int64
	UserId          int64
	Alert           string
	Lang            map[string]string
	CountSignArr    []int
	AutoPayments    []*autoPayment
	CurrencyList    map[int64]string
//	LastTxFormatted string
}

type autoPayment struct {
	Id, Currency_id, Last_payment_time, Recipient, Period, Sender int64
	Commission, Amount float64
}

func (c *Controller) AutoPayments() (string, error) {

	txType := "AutoPayments"
	txTypeId := utils.TypeInt(txType)
	timeNow := time.Now().Unix()

	var autoPayments []*autoPayment

	rows, err := c.Query(c.FormatQuery("SELECT id, amount, commission, currency_id, last_payment_time, period, sender, recipient FROM auto_payments WHERE sender = ? AND del_block_id = 0"), c.SessUserId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var amount, commission float64
		var id, currency_id, last_payment_time, period, sender, recipient int64
		err = rows.Scan(&id, &amount, &commission, &currency_id, &last_payment_time, &period, &sender, &recipient)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		auto_ := &autoPayment{Id : id, Amount : amount, Commission: commission, Currency_id: currency_id, Last_payment_time: last_payment_time, Period: period, Sender: sender, Recipient: recipient}
		autoPayments = append(autoPayments, auto_)
	}

/*	last_tx, err := c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"AutoPayments", "DelAutoPayment", "NewAutoPayment"}), 3, c.TimeFormat)
	lastTxFormatted := ""
	if len(last_tx) > 0 {
		lastTxFormatted, _ = utils.MakeLastTx(last_tx, c.Lang)
	}*/

	TemplateStr, err := makeTemplate("auto_payments", "AutoPayments", &AutoPaymentsPage{
		Alert:           c.Alert,
		Lang:            c.Lang,
		CountSignArr:    c.CountSignArr,
		ShowSignData:    c.ShowSignData,
		UserId:          c.SessUserId,
		TimeNow:         timeNow,
		TxType:          txType,
		TxTypeId:        txTypeId,
		SignData:        "",
		CurrencyList:    c.CurrencyListCf,
		AutoPayments: autoPayments,
		//LastTxFormatted: lastTxFormatted
		})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
