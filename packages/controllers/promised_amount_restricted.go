package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type promisedAmountRestricted struct {
	Alert           string
	SignData        string
	ShowSignData    bool
	CountSignArr    []int
	UserId          int64
	TxType           string
	TxTypeId         int64
	TimeNow          int64
	Lang            map[string]string
	Started         bool
	MinerId         int64
}

func (c *Controller) PromisedAmountRestricted() (string, error) {

	txType := "NewRestrictedPromisedAmount"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	var ( last_tx []map[string]string
		  started bool
		  err     error
	)
	if last_tx, err = c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"NewRestrictedPromisedAmount"}), 
		                    1, c.TimeFormat); err == nil && len(last_tx) > 0 {
		started = true
	}
	minerId,err := c.Single("SELECT miner_id FROM miners_data WHERE user_id  =  ?", c.SessUserId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}


	TemplateStr, err := makeTemplate("promised_amount_restricted", "PromisedAmountRestricted", &promisedAmountRestricted{
		Alert:           c.Alert,
		Started:         started,
		Lang:            c.Lang,
		CountSignArr:    c.CountSignArr,
		ShowSignData:    c.ShowSignData,
		TimeNow:          timeNow,
		TxType:           txType,
		TxTypeId:         txTypeId,
		MinerId:          minerId,
		UserId:          c.SessUserId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
