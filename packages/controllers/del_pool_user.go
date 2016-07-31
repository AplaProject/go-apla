package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

type DelPoolUserPage struct {
	SignData     string
	ShowSignData bool
	TxType       string
	TxTypeId     int64
	TimeNow      int64
	UserId       int64
	Alert        string
	Lang         map[string]string
	CountSignArr []int
	DelUserId     int64
}

func (c *Controller) DelPoolUser() (string, error) {

	txType := "DelPoolUser"
	txTypeId := utils.TypeInt(txType)
	timeNow := time.Now().Unix()

	delUserId := int64(utils.StrToFloat64(c.Parameters["del_user_id"]))

	TemplateStr, err := makeTemplate("del_pool_user", "delPoolUser", &DelPoolUserPage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		CountSignArr: c.CountSignArr,
		ShowSignData: c.ShowSignData,
		UserId:       c.SessUserId,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId,
		SignData:     "",
		DelUserId:     delUserId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
