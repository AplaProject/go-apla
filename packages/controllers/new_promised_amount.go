package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"net"
	"strings"
	"time"
)

type newPromisedAmountPage struct {
	Alert              string
	SignData           string
	ShowSignData       bool
	TxType             string
	TxTypeId           int64
	TimeNow            int64
	UserId             int64
	Lang               map[string]string
	CountSignArr       []int
	LastTxFormatted    string
	ConfigCommission   map[int64][]float64
	Navigate           string
	CurrencyId         int64
	CurrencyList       map[int64]map[string]string
	CurrencyListName   map[int64]string
	MaxPromisedAmounts map[string]string
	LimitsText         string
	PaymentSystems     map[string]string
	CountPs            []int
	Mobile             bool
	Mode               string
	IncNavigate        string
}

func (c *Controller) NewPromisedAmount() (string, error) {

	txType := "NewPromisedAmount"
	txTypeId := utils.TypeInt(txType)
	timeNow := time.Now().Unix()

	navigate := "promisedAmountList"
	if len(c.Navigate) > 0 {
		navigate = c.Navigate
	}

	rows, err := c.Query(c.FormatQuery(`
		SELECT id,
					 name,
					 full_name,
					 max_other_currencies
		FROM currency WHERE id != 1
		ORDER BY full_name`))
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	currencyList := make(map[int64]map[string]string)
	currencyListName := make(map[int64]string)
	defer rows.Close()
	for rows.Next() {
		var id int64
		var name, full_name, max_other_currencies string
		err = rows.Scan(&id, &name, &full_name, &max_other_currencies)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		currencyList[id] = map[string]string{"id": utils.Int64ToStr(id), "name": name, "full_name": full_name, "max_other_currencies": max_other_currencies}
		currencyListName[id] = name
	}

	paymentSystems, err := c.GetPaymentSystems()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	maxPromisedAmounts, err := c.GetMap(`SELECT currency_id, amount FROM max_promised_amounts WHERE block_id = 1`, "currency_id", "amount")
	maxPromisedAmountsMaxBlock, err := c.GetMap(`SELECT currency_id, amount FROM max_promised_amounts WHERE block_id = (SELECT max(block_id) FROM max_promised_amounts ) OR block_id = 0`, "currency_id", "amount")
	for k, v := range maxPromisedAmountsMaxBlock {
		maxPromisedAmounts[k] = v
	}
	for k, v := range maxPromisedAmounts {
		if countMiners,err := c.Single("SELECT count(id) FROM promised_amount where currency_id = ? AND status='mining'", k ).Int64(); err == nil {
			if countMiners < 1000 && utils.StrToFloat64(v) > float64(consts.MaxGreen[utils.StrToInt64(k)]) {
				maxPromisedAmounts[k] = utils.Int64ToStr(consts.MaxGreen[utils.StrToInt64(k)])
			}
		}
	}

	// валюта, которая выбрана в селект-боксе
	currencyId := int64(72)

	limitsText := strings.Replace(c.Lang["limits_text"], "[limit]", utils.Int64ToStr(c.Variables.Int64["limit_promised_amount"]), -1)
	limitsText = strings.Replace(limitsText, "[period]", c.Periods[c.Variables.Int64["limit_promised_amount_period"]], -1)

	countPs := []int{1}//, 2, 3, 4, 5}

	tcpHostPort, err := c.Single(`SELECT CASE WHEN m.pool_user_id > 0 then (SELECT tcp_host FROM miners_data WHERE user_id = m.pool_user_id) ELSE tcp_host END as tcp_host FROM miners_data as m WHERE m.user_id = ?`, c.SessUserId).String()
	tcpHost, _, _ := net.SplitHostPort(tcpHostPort)
	nodeIp, err := net.ResolveIPAddr("ip4", tcpHost)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	myIp, err := utils.GetHttpTextAnswer("http://api.ipify.org")
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	mode := "normal"
	if myIp != nodeIp.String() && len(tcpHost) > 0 {
		mode = "pool"
	}
//	fmt.Println(nodeIp.String(), myIp)

	TemplateStr, err := makeTemplate("new_promised_amount", "newPromisedAmount", &newPromisedAmountPage{
		Alert:              c.Alert,
		Lang:               c.Lang,
		CountSignArr:       c.CountSignArr,
		ShowSignData:       c.ShowSignData,
		UserId:             c.SessUserId,
		TimeNow:            timeNow,
		TxType:             txType,
		TxTypeId:           txTypeId,
		SignData:           "",
		ConfigCommission:   c.ConfigCommission,
		Navigate:           navigate,
		IncNavigate:        c.Navigate,
		CurrencyId:         currencyId,
		CurrencyList:       currencyList,
		CurrencyListName:   currencyListName,
		MaxPromisedAmounts: maxPromisedAmounts,
		LimitsText:         limitsText,
		PaymentSystems:     paymentSystems,
		Mobile:             utils.Mobile(),
		Mode:               mode,
		CountPs:            countPs})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
