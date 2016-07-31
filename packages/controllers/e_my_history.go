package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	//"fmt"
)

type eMyHistoryPage struct {
	Lang         map[string]string
	CurrencyList map[int64]string
	Commission   string
	UserId       int64
	MyHistory    []*EmyHistory
}

func (c *Controller) EMyHistory() (string, error) {

	var err error

	if c.SessUserId == 0 {
		return `<script language="javascript"> window.location.href = "` + c.EURL + `"</script>If you are not redirected automatically, follow the <a href="` + c.EURL + `">` + c.EURL + `</a>`, nil
	}

	currencyList, err := utils.EGetCurrencyList()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	//fmt.Println("currencyList", currencyList)
	var myHistory []*EmyHistory

	rows, err := c.Query(c.FormatQuery(`
			SELECT id, time, amount, sell_rate, sell_currency_id, buy_currency_id
			FROM e_trade
			WHERE user_id = ?
			ORDER BY time DESC
			LIMIT 40
			`), c.SessUserId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		myHist := new(EmyHistory)
		err = rows.Scan(&myHist.Id, &myHist.Time, &myHist.Amount, &myHist.SellRate, &myHist.SellCurrencyId, &myHist.BuyCurrencyId)
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		// определим тип ордера и пару
		if myHist.SellCurrencyId < 1000 {
			myHist.OrderType = "sell"
			myHist.SellRate = 1 / myHist.SellRate
			myHist.Total = myHist.Amount * myHist.SellRate
			myHist.Amount = myHist.Amount
			myHist.Pair = currencyList[myHist.SellCurrencyId] + "/" + currencyList[myHist.BuyCurrencyId]
		} else {
			myHist.OrderType = "buy"
			myHist.Total = myHist.Amount
			myHist.Amount = myHist.Amount * (1 / myHist.SellRate)
			myHist.Pair = currencyList[myHist.BuyCurrencyId] + "/" + currencyList[myHist.SellCurrencyId]
		}

		myHistory = append(myHistory, myHist)
	}

	TemplateStr, err := makeTemplate("e_my_history", "eMyHistory", &eMyHistoryPage{
		Lang:         c.Lang,
		UserId:       c.SessUserId,
		MyHistory:    myHistory,
		CurrencyList: currencyList})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}

type EmyHistory struct {
	Id, Time, SellCurrencyId, BuyCurrencyId int64
	Amount, SellRate, Total                 float64
	OrderType, Pair                         string
}
