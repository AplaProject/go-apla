package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/consts"
)

type promisedAmountRestrictedList struct {
	Alert           string
	SignData        string
	ShowSignData    bool
	CountSignArr    []int
	UserId          int64
	Pct float64
	Amount float64
	UserSn string
	LastTxQueueTx     bool
	LastTxTx          bool
	LastTxFormatted string
	MinerId         int64
	Attempts        int
	IsUpgrading     bool
	Lang            map[string]string
	MinWalletAmount string
}
/*
func (c *Controller) GetPromisedAmountCounter() ( float64, float64, error) {
	paRestricted, err := c.OneRow("SELECT * FROM promised_amount_restricted WHERE user_id = ?", c.SessUserId).String()
	if err != nil {
		return 0, 0, err
	}
	if _, ok := paRestricted[`user_id`]; !ok {
		return 0, 0, nil
	}
	
	amount := utils.StrToFloat64(paRestricted["amount"])
	// Временная проверка для старого формата таблицы promised_amount_restricted. 
	if _, ok := paRestricted["start_time"]; ok && utils.StrToInt64(paRestricted["last_update"]) == 0 {
		paRestricted["last_update"] = paRestricted["start_time"]
	}
	profit, err := c.CalcProfitGen(utils.StrToInt64(paRestricted["currency_id"]), amount, c.SessUserId, utils.StrToInt64(paRestricted["last_update"]), utils.Time(), "wallet")
	if err != nil {
		return 0, 0, err
	}
	profit += amount
	
	pct, err := c.Single(c.FormatQuery("SELECT user FROM pct WHERE currency_id  =  ? ORDER BY block_id DESC"), utils.StrToInt64(paRestricted["currency_id"])).Float64()
	if err != nil {
		return 0, 0, err
	}
	return profit, pct, nil
}*/

func (c *Controller) PromisedAmountRestrictedList() (string, error) {

	profit, pct, err := c.GetPromisedAmountCounter(c.SessUserId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	user, err := c.OneRow(`SELECT status, sn_attempts, sn_url_id FROM users WHERE user_id = ?`, c.SessUserId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	userSn := user[`status`]

	var lastTxQueueTx, lastTxTx bool
	lastTx, err := c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"UpgradeUser", "MiningSn"}), 1, c.TimeFormat)
	lastTxFormatted := ""
	if len(lastTx) > 0 {
		lastTxFormatted, _ = utils.MakeLastTx(lastTx, c.Lang)
		if len(lastTx[0]["queue_tx"]) > 0 {
			lastTxQueueTx = true
		}
		if len(lastTx[0]["tx"]) > 0 {
			lastTxTx = true
		}
	}
	var ( attempts int
		isUpgrading bool
	)
	attempts = consts.SN_USER_ATTEMPTS - utils.StrToInt( user[`sn_attempts`] )
	if attempts < 0 {
		attempts = 0
	}
	if userSn != "sn_user" {
		// Есть ли в процессе транзакция UpgradeUser
		for _, tx := range lastTx {
			if consts.TxTypes[ utils.StrToInt(tx[`type`])] == "UpgradeUser" &&
			    utils.StrToInt( tx[`block_id`] ) == 0 && len(tx[`txerror`]) == 0 {
				isUpgrading = true
			}
		}
		if userSn == "user" && len(user[`sn_url_id`]) > 0 {
			// идет проверка соц аккаунта
			isUpgrading = true
		}
	}

	
	minerId, err := c.Single("SELECT miner_id FROM miners_data WHERE user_id  =  ?", c.SessUserId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	minWalletAmount, err := c.Single("SELECT amount FROM wallets WHERE user_id  =  ? and currency_id = ?", c.SessUserId, 72).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	TemplateStr, err := makeTemplate("promised_amount_restricted_list", "PromisedAmountRestrictedList", &promisedAmountRestrictedList{
		Alert:           c.Alert,
		Lang:            c.Lang,
		CountSignArr:    c.CountSignArr,
		Pct : pct,
		Amount : profit,
		LastTxFormatted: lastTxFormatted,
		LastTxQueueTx:     lastTxQueueTx,
		LastTxTx:          lastTxTx,
		ShowSignData:    c.ShowSignData,
		MinerId:         minerId,
		MinWalletAmount: minWalletAmount,
		UserSn:          userSn,
		Attempts:        attempts,
		IsUpgrading:     isUpgrading,
		UserId:          c.SessUserId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
