package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

type changeCommissionPage struct {
	Alert            string
	SignData         string
	ShowSignData     bool
	TxType           string
	TxTypeId         int64
	TimeNow          int64
	UserId           int64
	Lang             map[string]string
	CountSignArr     []int
//	LastTxFormatted  string
	CurrencyList     map[int64]string
	ConfigCommission map[int64][]float64
	Navigate         string
	Commission       map[int64][]float64
}

func (c *Controller) ChangeCommission() (string, error) {

	txType := "ChangeCommission"
	txTypeId := utils.TypeInt(txType)
	timeNow := time.Now().Unix()

	navigate := "changeCommission"
	if len(c.Navigate) > 0 {
		navigate = c.Navigate
	}

	minCommission := map[string]string{"WOC": "0.01", "AED": "0.04", "AOA": "0.96", "ARS": "0.06", "AUD": "0.01", "AZN": "0.01", "BDT": "0.78", "BGN": "0.01", "BOB": "0.07", "BRL": "0.02", "BYR": "89.25", "CAD": "0.01", "CHF": "0.01", "CLP": "5.13", "CNY": "0.06", "COP": "19.11", "CRC": "4.98", "CZK": "0.19", "DKK": "0.06", "DOP": "0.42", "DZD": "0.81", "EGP": "0.07", "EUR": "0.01", "GBP": "0.01", "GEL": "0.02", "GHS": "0.02", "GTQ": "0.08", "HKD": "0.08", "HRK": "0.06", "HUF": "2.25", "IDR": "103.85", "ILS": "0.04", "INR": "0.62", "IQD": "11.64", "IRR": "999.99", "JOD": "0.01", "JPY": "0.98", "KES": "0.87", "KRW": "11.14", "KWD": "0.01", "KZT": "1.53", "LBP": "15.11", "LKR": "1.32", "MAD": "0.08", "MXN": "0.13", "MYR": "0.03", "NGN": "1.61", "NOK": "0.06", "NPR": "0.98", "NZD": "0.01", "PEN": "0.03", "PHP": "0.44", "PKR": "1.03", "PLN": "0.03", "QAR": "0.04", "RON": "0.03", "RSD": "0.85", "RUB": "0.33", "SAR": "0.04", "SDG": "0.04", "SEK": "0.07", "SGD": "0.01", "SVC": "0.09", "SYP": "1.08", "THB": "0.31", "TND": "0.02", "TRY": "0.02", "TWD": "0.30", "TZS": "16.19", "UAH": "0.08", "UGX": "25.79", "USD": "0.01", "UZS": "21.15", "VEF": "0.06", "VND": "210.95", "YER": "2.15", "ZAR": "0.10", "BTC": "0.01", "LTC": "0.01"}

	currencyMin := make(map[int64]string)
	for id, name := range c.CurrencyList {
		currencyMin[id] = minCommission[name]
	}

	myCommission := make(map[int64][]float64)
	if c.SessRestricted == 0 {
		rows, err := c.Query(c.FormatQuery("SELECT currency_id, pct, min, max FROM " + c.MyPrefix + "my_commission"))
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		defer rows.Close()
		for rows.Next() {
			var pct, min, max float64
			var currency_id int64
			err = rows.Scan(&currency_id, &pct, &min, &max)
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			myCommission[currency_id] = []float64{pct, min, max}
		}
	}

	currencyList := c.CurrencyList
	commission := make(map[int64][]float64)
	for currency_id, _ := range currencyList {
		if len(myCommission[currency_id]) > 0 {
			commission[currency_id] = myCommission[currency_id]
		} else {
			commission[currency_id] = []float64{0.1, utils.StrToFloat64(currencyMin[currency_id]), 0}
		}
	}

	// для CF-проектов
	currencyList[1000] = "Crowdfunding"
	if len(myCommission[1000]) > 0 {
		commission[1000] = myCommission[1000]
	} else {
		commission[1000] = []float64{0.1, 0.01, 0}
	}

/*	last_tx, err := c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"ChangeArbitratorConditions", "MoneyBack"}), 3, c.TimeFormat)
	lastTxFormatted := ""
	if len(last_tx) > 0 {
		lastTxFormatted, _ = utils.MakeLastTx(last_tx, c.Lang)
	}*/

	//limitsText := strings.Replace(c.Lang["change_commission_limits_text"], "[limit]", utils.Int64ToStr(c.Variables.Int64["limit_commission"]), -1)
	//limitsText = strings.Replace(limitsText, "[period]", c.Periods[c.Variables.Int64["limit_commission_period"]], -1)

	log.Debug("commission:", commission)

	TemplateStr, err := makeTemplate("change_commission", "changeCommission", &changeCommissionPage{
		Alert:            c.Alert,
		Lang:             c.Lang,
		CountSignArr:     c.CountSignArr,
		ShowSignData:     c.ShowSignData,
		UserId:           c.SessUserId,
		TimeNow:          timeNow,
		TxType:           txType,
		TxTypeId:         txTypeId,
		SignData:         "",
//		LastTxFormatted:  lastTxFormatted,
		CurrencyList:     c.CurrencyList,
		ConfigCommission: c.ConfigCommission,
		Navigate:         navigate,
		Commission:       commission})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
