package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type changeKeyRequestPage struct {
	Alert        string
	SignData     string
	ShowSignData bool
	CountSignArr []int
	Lang         map[string]string
	UserId       int64
	TxType       string
	TxTypeId     int64
	TimeNow      int64
}

func (c *Controller) ChangeKeyRequest() (string, error) {

	txType := "ChangeKeyActive"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	TemplateStr, err := makeTemplate("change_key_request", "changeKeyRequest", &changeKeyRequestPage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		ShowSignData: c.ShowSignData,
		SignData:     "",
		UserId:       c.SessUserId,
		CountSignArr: c.CountSignArr,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
