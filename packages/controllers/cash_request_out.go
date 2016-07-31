package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"

	"fmt"
)

type cashRequestOutPage struct {
	Alert               string
	SignData            string
	ShowSignData        bool
	TxType              string
	TxTypeId            int64
	TimeNow             int64
	UserId              int64
	Lang                map[string]string
	CountSignArr        []int
	CurrencyList        map[int64]string
	CashRequestsStatus  map[string]string
	JsonCurrencyWallets string
	PaymentSystems      map[string]string
	MinPromisedAmount   int64
	MaxLength           int
	AvailableCurrency   []int64
	MyCashRequests      []map[string]string
	Code                string
	HashCode            string
}

func (c *Controller) CashRequestOut() (string, error) {

	txType := "CashRequestOut"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	cashRequestsStatus := map[string]string{"my_pending": c.Lang["local_pending"], "pending": c.Lang["pending"], "approved": c.Lang["approved"], "rejected": c.Lang["rejected"]}

	jsonCurrencyWallets := ""
	var availableCurrency []int64
	// список отравленных нами запросов
	myCashRequests, err := c.GetAll("SELECT * FROM "+c.MyPrefix+"my_cash_requests WHERE to_user_id != ?", -1, c.SessUserId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	// получаем список кошельков, на которых есть FC
	rows, err := c.Query(c.FormatQuery("SELECT amount, currency_id, last_update FROM wallets WHERE user_id = ? AND currency_id < 1000"), c.SessUserId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var amount float64
		var currency_id, last_update int64
		err = rows.Scan(&amount, &currency_id, &last_update)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if currency_id == 1 {
			continue
		}
		profit, err := c.CalcProfitGen(currency_id, amount, c.SessUserId, last_update, timeNow, "wallet")
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		amount += profit
		jsonCurrencyWallets += fmt.Sprintf(`"%d":["%s","%v"],`, currency_id, c.CurrencyList[currency_id], amount)
		availableCurrency = append(availableCurrency, currency_id)
	}
	jsonCurrencyWallets = jsonCurrencyWallets[:len(jsonCurrencyWallets)-1]
	//jsonCurrencyWallets = "\""
	log.Debug("jsonCurrencyWallets", jsonCurrencyWallets)
	code := utils.Md5(utils.RandSeq(50))
	hashCode := utils.DSha256(code)

	TemplateStr, err := makeTemplate("cash_request_out", "cashRequestOut", &cashRequestOutPage{
		Alert:               c.Alert,
		Lang:                c.Lang,
		CountSignArr:        c.CountSignArr,
		ShowSignData:        c.ShowSignData,
		UserId:              c.SessUserId,
		TimeNow:             timeNow,
		TxType:              txType,
		TxTypeId:            txTypeId,
		SignData:            "",
		CurrencyList:        c.CurrencyList,
		PaymentSystems:      c.PaymentSystems,
		JsonCurrencyWallets: jsonCurrencyWallets,
		CashRequestsStatus:  cashRequestsStatus,
		AvailableCurrency:   availableCurrency,
		MinPromisedAmount:   c.Variables.Int64["min_promised_amount"],
		MyCashRequests:      myCashRequests,
		Code:                string(code),
		HashCode:            string(hashCode),
		MaxLength:           200})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
