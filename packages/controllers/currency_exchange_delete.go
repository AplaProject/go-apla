package controllers

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

type currencyExchangeDeletePage struct {
	SignData     string
	ShowSignData bool
	TxType       string
	TxTypeId     int64
	TimeNow      int64
	UserId       int64
	Alert        string
	Lang         map[string]string
	CountSignArr []int
	DelId        int64
}

func (c *Controller) CurrencyExchangeDelete() (string, error) {

	txType := "DelForexOrder"
	txTypeId := utils.TypeInt(txType)
	timeNow := time.Now().Unix()

	delId := utils.StrToInt64(c.Parameters["del_id"])
	signData := fmt.Sprintf("%d,%d,%d,%d", txTypeId, timeNow, c.SessUserId, delId)

	/*data, err := static.Asset("static/templates/currency_exchange_delete.html")
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	signatures, err := static.Asset("static/templates/signatures.html")
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	alert_success, err := static.Asset("static/templates/alert_success.html")
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	t := template.Must(template.New("template").Parse(string(data)))
	t = template.Must(t.Parse(string(alert_success)))
	t = template.Must(t.Parse(string(signatures)))
	b := new(bytes.Buffer)
	t.ExecuteTemplate(b, "currencyExchangeDelete", &currencyExchangeDeletePage{
		Lang: c.Lang,
		CountSignArr: c.CountSignArr,
		ShowSignData: c.ShowSignData,
		UserId: c.SessUserId,
		TimeNow: timeNow,
		TxType: txType,
		TxTypeId: txTypeId,
		DelId: delId,
		SignData: signData})
	return b.String(), nil*/

	TemplateStr, err := makeTemplate("currency_exchange_delete", "currencyExchangeDelete", &currencyExchangeDeletePage{
		Lang:         c.Lang,
		CountSignArr: c.CountSignArr,
		ShowSignData: c.ShowSignData,
		UserId:       c.SessUserId,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId,
		DelId:        delId,
		SignData:     signData})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
