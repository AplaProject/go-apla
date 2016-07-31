package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"strings"
)

type changePrimaryKeyPage struct {
	Alert             string
	SignData          string
	ShowSignData      bool
	CountSignArr      []int
	Lang              map[string]string
	LimitsText        string
	LastTxFormatted   string
	LastChangeKeyTime int64
	LastTxQueueTx     bool
	LastTxTx          bool
	UserId            int64
	LastTx            []map[string]string
	MyKeys            []map[string]string
	StatusArray       map[string]string
	TxType            string
	TxTypeId          int64
	TimeNow           int64
	IOS bool
	Android bool
	Mobile bool
}

func (c *Controller) ChangePrimaryKey() (string, error) {

	var err error

	txType := "ChangePrimaryKey"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	var myKeys []map[string]string
	if c.SessRestricted == 0 {
		myKeys, err = c.GetAll(`SELECT * FROM `+c.MyPrefix+`my_keys ORDER BY id DESC`, -1)
	}

	statusArray := map[string]string{"my_pending": c.Lang["local_pending"], "approved": c.Lang["status_approved"]}

	// узнаем, когда последний раз была смена ключа, чтобы не показывать юзеру страницу смены
	lastChangeKeyTime, err := c.Single("SELECT time FROM log_time_primary_key WHERE user_id  =  ? ORDER BY time DESC", c.SessUserId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	limitsText := strings.Replace(c.Lang["change_primary_key_limits_text"], "[limit]", utils.Int64ToStr(c.Variables.Int64["limit_primary_key"]), -1)
	limitsText = strings.Replace(limitsText, "[period]", c.Periods[c.Variables.Int64["limit_primary_key_period"]], -1)

	var lastTxQueueTx, lastTxTx bool
	lastTx, err := c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"ChangePrimaryKey"}), 1, c.TimeFormat)
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

	TemplateStr, err := makeTemplate("change_primary_key", "changePrimaryKey", &changePrimaryKeyPage{
		Alert:             c.Alert,
		Lang:              c.Lang,
		ShowSignData:      c.ShowSignData,
		SignData:          "",
		UserId:            c.SessUserId,
		CountSignArr:      c.CountSignArr,
		LimitsText:        limitsText,
		LastTxQueueTx:     lastTxQueueTx,
		LastTxTx:          lastTxTx,
		LastTxFormatted:   lastTxFormatted,
		LastChangeKeyTime: lastChangeKeyTime,
		LastTx:            lastTx,
		MyKeys:            myKeys,
		StatusArray:       statusArray,
		TimeNow:           timeNow,
		TxType:            txType,
		IOS: utils.IOS(),
		Android: utils.Android(),
		Mobile: utils.Mobile(),
		TxTypeId:          txTypeId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
