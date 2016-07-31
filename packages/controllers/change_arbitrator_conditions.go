package controllers

import (
	"encoding/json"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

type changeArbitratorConditionsPage struct {
	SignData        string
	ShowSignData    bool
	TxType          string
	TxTypeId        int64
	TimeNow         int64
	UserId          int64
	Alert           string
	Lang            map[string]string
	CountSignArr    []int
	CurrencyList    map[int64]string
	PendingTx       int64
	LastTxFormatted string
	Conditions      map[int64][5]string
	Commission      map[int64][]float64
}

func (c *Controller) ChangeArbitratorConditions() (string, error) {

	txType := "ChangeArbitratorConditions"
	txTypeId := utils.TypeInt(txType)
	timeNow := time.Now().Unix()

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
			commission[currency_id] = []float64{0.1, 0.01, 0}
		}
	}

	// для CF-проектов
	currencyList[1000] = "Crowdfunding"
	if len(myCommission[1000]) > 0 {
		commission[1000] = myCommission[1000]
	} else {
		commission[1000] = []float64{0.1, 0.01, 0}
	}
	arbitratorConditionsJson, err := c.Single("SELECT conditions FROM arbitrator_conditions WHERE user_id  =  ?", c.SessUserId).Bytes()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	arbitratorConditionsMap_ := make(map[string][5]string)
	if len(arbitratorConditionsJson) > 0 {
		err = json.Unmarshal(arbitratorConditionsJson, &arbitratorConditionsMap_)
		// арбитр к этому моменту мог передумать и убрать свои условия, уйдя из арбитров для новых сделок поставив [0] что вызовет тут ошибку
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}
	arbitratorConditionsMap := make(map[int64][5]string)
	for k, v := range arbitratorConditionsMap_ {
		arbitratorConditionsMap[utils.StrToInt64(k)] = v
	}
	if len(arbitratorConditionsMap) == 0 {
		arbitratorConditionsMap[72] = [5]string{"0.01", "0", "0.01", "0", "0.1"}
		arbitratorConditionsMap[23] = [5]string{"0.01", "0", "0.01", "0", "0.1"}
	}

	last_tx, err := c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"ChangeArbitratorConditions"}), 3, c.TimeFormat)
	lastTxFormatted := ""
	var pendingTx_ map[int64]int64
	if len(last_tx) > 0 {
		lastTxFormatted, pendingTx_ = utils.MakeLastTx(last_tx, c.Lang)
	}
	pendingTx := pendingTx_[txTypeId]

	TemplateStr, err := makeTemplate("change_arbitrator_conditions", "changeArbitratorConditions", &changeArbitratorConditionsPage{
		Alert:           c.Alert,
		Lang:            c.Lang,
		CountSignArr:    c.CountSignArr,
		ShowSignData:    c.ShowSignData,
		UserId:          c.SessUserId,
		TimeNow:         timeNow,
		TxType:          txType,
		TxTypeId:        txTypeId,
		SignData:        "",
		CurrencyList:    c.CurrencyList,
		PendingTx:       pendingTx,
		LastTxFormatted: lastTxFormatted,
		Conditions:      arbitratorConditionsMap,
		Commission:      commission})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
