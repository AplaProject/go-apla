package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"strings"
	"time"
)

type promisedAmountListPage struct {
	Alert                        string
	SignData                     string
	ShowSignData                 bool
	TxType                       string
	TxTypeId                     int64
	TimeNow                      int64
	UserId                       int64
	Lang                         map[string]string
	CountSignArr                 []int
	LastTxFormatted              string
	CurrencyList                 map[int64]string
	ConfigCommission             map[int64][]float64
	Navigate                     string
	Commission                   map[int64][]float64
	PromisedAmountListAccepted   []utils.PromisedAmounts
	ActualizationPromisedAmounts int64
	LimitsText                   string
	DisableNewAmount             bool
}

func (c *Controller) PromisedAmountList() (string, error) {
	var disableNewAmount bool
	txType := "PromisedAmount"
	txTypeId := utils.TypeInt(txType)
	timeNow := time.Now().Unix()

	last_tx, err := c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"NewPromisedAmount", "ChangePromisedAmount", "DelPromisedAmount", "ForRepaidFix", "ActualizationPromisedAmounts", "Mining"}), 3, c.TimeFormat)
	lastTxFormatted := ""
	if len(last_tx) > 0 {
		lastTxFormatted, _ = utils.MakeLastTx(last_tx, c.Lang)
	}

	limitsText := strings.Replace(c.Lang["change_commission_limits_text"], "[limit]", utils.Int64ToStr(c.Variables.Int64["limit_promised_amount"]), -1)
	limitsText = strings.Replace(limitsText, "[period]", c.Periods[c.Variables.Int64["limit_promised_amount_period"]], -1)

	actualizationPromisedAmounts, promisedAmountListAccepted, _, err := c.GetPromisedAmounts(c.SessUserId, c.Variables.Int64["cash_request_time"])
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	for _, tx := range last_tx {
		if utils.StrToInt64( tx[`block_id`] ) == 0 {
			if len( tx[`tx`] ) > 0 || len( tx[`queue_tx`] ) > 0 {
				// Есть необработанные транзакции
				disableNewAmount = true
			}
			IDB:
			for _, idb := range []string{`queue_tx`,`transactions`}{
				data, err := c.Single(`SELECT data FROM `+ idb +` WHERE hex(hash)=?`, utils.BinToHex([]byte(tx["hash"])) ).Bytes(); 
				if err==nil && len( data ) > 0 {
					data2 := data[5:]			
					length := utils.DecodeLength(&data2)
					utils.BytesShift(&data2, length)
					length = utils.DecodeLength(&data2)
					idPromise := utils.StrToInt64( string(utils.BytesShift(&data2, length)))
					for i, ipromise := range promisedAmountListAccepted {
						if ipromise.Id == idPromise {
							promisedAmountListAccepted[i].InProcess = true
							break IDB
						}
					}
				}
			}
		}
	}
	for i, ipromise := range promisedAmountListAccepted {
		if ipromise.CurrencyId > 1 {
			if countMiners,err := c.Single("SELECT count(id) FROM promised_amount where currency_id = ? AND status='mining'", 
			               ipromise.CurrencyId ).Int64(); err == nil {
				if countMiners < 1000 && ipromise.MaxAmount > float64(consts.MaxGreen[ipromise.CurrencyId]) {
					promisedAmountListAccepted[i].MaxAmount = float64(consts.MaxGreen[ipromise.CurrencyId])
				}
			}
		}
	}
	if !disableNewAmount {
		// Сразу будем скрывать кнопку если обещанная сумма уже есть, но еще не одобрена,
		// а также есди у какой-то валюты уже достигнут предел в max_other_currencies
		existsCurrencies, err := c.DCDB.GetAll(`SELECT currency_id, c.max_other_currencies FROM promised_amount 
			LEFT JOIN currency as c ON c.id = currency_id
			WHERE user_id = ? AND del_block_id  =  0 AND del_mining_block_id  =  0 GROUP BY currency_id`, -1, c.SessUserId )
		if err == nil && len(existsCurrencies) > 0 {
			woc := false
			for _, item := range existsCurrencies {
				if utils.StrToInt(item[`currency_id`]) == 1 {
					woc = true
					continue
				}
				if len(existsCurrencies) > utils.StrToInt(item[`max_other_currencies`]) {
					disableNewAmount = true
				}
			}
			if !woc {
				disableNewAmount = true
			}
		}
	}

	TemplateStr, err := makeTemplate("promised_amount_list", "promisedAmountList", &promisedAmountListPage{
		Alert:                        c.Alert,
		Lang:                         c.Lang,
		CountSignArr:                 c.CountSignArr,
		ShowSignData:                 c.ShowSignData,
		UserId:                       c.SessUserId,
		TimeNow:                      timeNow,
		TxType:                       txType,
		TxTypeId:                     txTypeId,
		SignData:                     "",
		LastTxFormatted:              lastTxFormatted,
		CurrencyList:                 c.CurrencyList,
		PromisedAmountListAccepted:   promisedAmountListAccepted,
		ActualizationPromisedAmounts: actualizationPromisedAmounts,
		DisableNewAmount:             disableNewAmount,
		LimitsText:                   limitsText})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
