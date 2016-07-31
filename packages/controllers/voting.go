package controllers

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"strings"
	"github.com/DayLightProject/go-daylight/packages/consts"
)

type VotingPage struct {
	SignData                   string
	ShowSignData               bool
	TxType                     string
	TxTypeId                   int64
	TimeNow                    int64
	UserId                     int64
	Alert                      string
	Lang                       map[string]string
	CountSignArr               []int
	PromisedAmountCurrencyList map[int64]map[string]string
	MaxOtherCurrenciesCount    []int
	RefsNums                   []int
	Refs                       []string
	Referral                   map[string]int64
	MinerNewbie                string
	MaxCurrencyId              int64
	AllMaxPromisedAmount       []int64
	AllPct                     [391]map[string]string
//	LastTxFormatted            string
	WaitVoting                 map[int64]string
	CurrencyList               map[int64]string
	JsPct                      string
	MaxPromisedAmountSelectBox map[int64]string
	MinerPctSelectBox			map[int64]string
	UserPctSelectBox			map[int64]string
	MaxOtherCurrenciesSelectBox			map[int64]string

}

func (c *Controller) Voting() (string, error) {

	txType := "VotesComplex"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	waitVoting := make(map[int64]string)
	promisedAmountCurrencyList := make(map[int64]map[string]string)

	// голосовать майнер может только после того, как пройдет  miner_newbie_time сек
	regTime, err := c.Single("SELECT reg_time FROM miners_data WHERE user_id  =  ?", c.SessUserId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	// ****************************************** убрать
	c.Variables.Int64["min_miners_of_voting"] = 1;
	// ****************************************** убрать

	minerNewbie := ""
	if regTime > utils.Time()-c.Variables.Int64["miner_newbie_time"] && c.SessUserId != 1 {
		minerNewbie = strings.Replace(c.Lang["hold_time_wait2"], "[sec]", utils.TimeLeft(c.Variables.Int64["miner_newbie_time"]-(utils.Time()-regTime), c.Lang), -1)
	} else {
		// валюты
		rows, err := c.Query(c.FormatQuery(`
				SELECT currency_id,
							  name,
							  full_name,
							  start_time
				FROM promised_amount
					LEFT JOIN currency ON currency.id = promised_amount.currency_id
				WHERE user_id = ? AND
							 status IN ('mining', 'repaid') AND
							 start_time > 0 AND
							 del_block_id = 0
				GROUP BY currency_id
				`), c.SessUserId)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		defer rows.Close()
		for rows.Next() {
			var currency_id, start_time int64
			var name, full_name string
			err = rows.Scan(&currency_id, &name, &full_name, &start_time)
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			// после добавления обещанной суммы должно пройти не менее min_hold_time_promise_amount сек, чтобы за неё можно было голосовать
			if start_time > utils.Time()-c.Variables.Int64["min_hold_time_promise_amount"] {
				waitVoting[currency_id] = strings.Replace(c.Lang["hold_time_wait"], "[sec]", utils.TimeLeft(c.Variables.Int64["min_hold_time_promise_amount"]-(utils.Time()-start_time), c.Lang), -1)
				continue
			}
			// если по данной валюте еще не набралось >1000 майнеров, то за неё голосовать нельзя.
			countMiners, err := c.Single(`
					SELECT count(user_id)
					FROM promised_amount
					WHERE start_time < ? AND
								 del_block_id = 0 AND
								 status IN ('mining', 'repaid') AND
								 currency_id = ? AND
								 del_block_id = 0
					GROUP BY  user_id
					`, utils.Time()-c.Variables.Int64["min_hold_time_promise_amount"], currency_id).Int64()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			if countMiners < c.Variables.Int64["min_miners_of_voting"] {
				waitVoting[currency_id] = strings.Replace(c.Lang["min_miners_count"], "[miners_count]", utils.Int64ToStr(c.Variables.Int64["min_miners_of_voting"]), -1)
				waitVoting[currency_id] = strings.Replace(waitVoting[currency_id], "[remaining]", utils.Int64ToStr(c.Variables.Int64["min_miners_of_voting"]-countMiners), -1)
				continue
			}
			// голосовать можно не чаще 1 раза в 2 недели
			voteTime, err := c.Single("SELECT time FROM log_time_votes_complex WHERE user_id  =  ? AND time > ?", c.SessUserId, utils.Time()-c.Variables.Int64["limit_votes_complex_period"]).Int64()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			if voteTime > 0 {
				waitVoting[currency_id] = strings.Replace(c.Lang["wait_voting"], "[sec]", utils.TimeLeft(c.Variables.Int64["limit_votes_complex_period"]-(utils.Time()-voteTime), c.Lang), -1)
				continue
			}

			// получим наши предыдущие голоса
			votesUserPct, err := c.Single("SELECT pct FROM votes_user_pct WHERE user_id  =  ? AND currency_id  =  ?", c.SessUserId, currency_id).String()
			if err != nil {
				return "", utils.ErrInfo(err)
			}

			votesMinerPct, err := c.Single("SELECT pct FROM votes_miner_pct WHERE user_id  =  ? AND currency_id  =  ?", c.SessUserId, currency_id).String()
			if err != nil {
				return "", utils.ErrInfo(err)
			}

			votesMaxOtherCurrencies, err := c.Single("SELECT count FROM votes_max_other_currencies WHERE user_id  =  ? AND currency_id  =  ?", c.SessUserId, currency_id).Int64()
			if err != nil {
				return "", utils.ErrInfo(err)
			}

			votesMaxPromisedAmount, err := c.Single("SELECT amount FROM votes_max_promised_amount WHERE user_id  =  ? AND currency_id  =  ?", c.SessUserId, currency_id).Int64()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			promisedAmountCurrencyList[currency_id] = make(map[string]string)
			promisedAmountCurrencyList[currency_id]["votes_user_pct"] = votesUserPct
			promisedAmountCurrencyList[currency_id]["votes_miner_pct"] = votesMinerPct
			promisedAmountCurrencyList[currency_id]["votes_max_other_currencies"] = utils.Int64ToStr(votesMaxOtherCurrencies)
			promisedAmountCurrencyList[currency_id]["votes_max_promised_amount"] = utils.Int64ToStr(votesMaxPromisedAmount)
			promisedAmountCurrencyList[currency_id]["name"] = name
		}
	}

	referral, err := c.OneRow("SELECT first, second, third FROM votes_referral WHERE user_id  =  ?", c.SessUserId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if len(referral) == 0 {
		referral["first"] = int64(utils.RandInt(0, 30))
		referral["second"] = int64(utils.RandInt(0, 30))
		referral["third"] = int64(utils.RandInt(0, 30))
	}

	maxCurrencyId, err := c.Single("SELECT max(id) FROM currency").Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	allMaxPromisedAmount := []int64{1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000, 10000, 20000, 50000, 100000, 200000, 500000, 1000000, 2000000, 5000000, 10000000, 20000000, 50000000, 100000000, 200000000, 500000000, 1000000000}

	allPct := utils.GetPctArray()
	pctArray := utils.GetPctArray()
	jsPct := "{"
	for year, sec := range pctArray {
		jsPct += fmt.Sprintf(`%v: '%v',`, year, sec)
	}
	jsPct = jsPct[:len(jsPct)-1] + "}"

/*	lastTx, err := c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"VotesComplex"}), 1, c.TimeFormat)
	lastTxFormatted := ""
	if len(lastTx) > 0 {
		lastTxFormatted, _ = utils.MakeLastTx(lastTx, c.Lang)
	}*/

	refs := []string{"first", "second", "third"}
	refsNums := []int{0, 5, 10, 15, 20, 25, 30}
	maxOtherCurrenciesCount:= []int{0, 1, 2, 3, 4}

	maxPromisedAmountSelectBox := make(map[int64]string)
	minerPctSelectBox := make(map[int64]string)
	userPctSelectBox := make(map[int64]string)
	maxOtherCurrenciesSelectBox := make(map[int64]string)
	for currencyId, data := range promisedAmountCurrencyList {
		selectBox:=""
		for _, amount := range allMaxPromisedAmount {
			sel := ""
			if data["votes_max_promised_amount"] ==  utils.Int64ToStr(amount) {
				sel = "selected"
			}
			color := ""
			if amount<=consts.MaxGreen[currencyId] {
				color = `style="background-color:#B1D253"`
			} else if amount<=consts.MaxGreen[currencyId]*10 {
				color = `style="background-color:#FFBA3F"`
			} else {
				color = `style="background-color:#EE3224"`
			}
			selectBox=selectBox+`<option `+sel+` `+color+`>`+utils.Int64ToStr(amount)+`</option>`
		}
		maxPromisedAmountSelectBox[currencyId] = selectBox

		selectBox=""
		for _, pctData := range allPct {
			for y, sec := range pctData {
				sel := ""
				if data["votes_miner_pct"] == sec {
					sel = "selected"
				}
				selectBox=selectBox+`<option  value="`+ sec +`" `+sel+`>`+utils.ClearNull(y, 2)+`</option>`
			}
		}
		minerPctSelectBox[currencyId] = selectBox

		selectBox=""
		for _, pctData := range allPct {
			for y, sec := range pctData {
				sel := ""
				if data["votes_user_pct"] == sec {
					sel = "selected"
				}
				selectBox=selectBox+`<option value="`+ sec +`" `+sel+`>`+utils.ClearNull(y, 2)+`</option>`
			}
		}
		userPctSelectBox[currencyId] = selectBox

		selectBox=""
		for _, v := range maxOtherCurrenciesCount {
			sel := ""
			if data["votes_max_other_currencies"] == utils.IntToStr(v) {
				sel = "selected"
			}
			selectBox=selectBox+`<option `+sel+`>`+utils.IntToStr(v)+`</option>`
		}
		maxOtherCurrenciesSelectBox[currencyId] = selectBox
	}

	TemplateStr, err := makeTemplate("voting", "voting", &VotingPage{
		Alert:                      c.Alert,
		Lang:                       c.Lang,
		CountSignArr:               c.CountSignArr,
		ShowSignData:               c.ShowSignData,
		UserId:                     c.SessUserId,
		TimeNow:                    timeNow,
		TxType:                     txType,
		TxTypeId:                   txTypeId,
		SignData:                   "",
		PromisedAmountCurrencyList: promisedAmountCurrencyList,
		MaxOtherCurrenciesCount:    []int{0, 1, 2, 3, 4},
		RefsNums:                   refsNums,
		Referral:                   referral,
		MinerNewbie:                minerNewbie,
		MaxCurrencyId:              maxCurrencyId,
		AllMaxPromisedAmount:       allMaxPromisedAmount,
		AllPct:                     allPct,
//		LastTxFormatted:            lastTxFormatted,
		WaitVoting:                 waitVoting,
		CurrencyList:               c.CurrencyList,
		JsPct:                      jsPct,
		MaxOtherCurrenciesSelectBox : maxOtherCurrenciesSelectBox,
		MaxPromisedAmountSelectBox : maxPromisedAmountSelectBox,
		MinerPctSelectBox : minerPctSelectBox,
		UserPctSelectBox : userPctSelectBox,
		Refs:                       refs})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
