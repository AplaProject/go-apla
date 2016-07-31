package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type upgradeUserPage struct {
	Alert        string
	SignData     string
	ShowSignData bool
	CountSignArr []int
	Lang         map[string]string
	UserId       int64
	LimitsText   string
	Community    bool
	TxType       string
	TxTypeId     int64
	TimeNow      int64
}

func (c *Controller) UpgradeUser() (string, error) {

	txType := "UpgradeUser"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	TemplateStr, err := makeTemplate("upgrade_user", "upgradeUser", &upgradeUserPage{
		Alert:        c.Alert,
		UserId:       c.SessUserId,
		CountSignArr: c.CountSignArr,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId,
		ShowSignData: c.ShowSignData,
		Community:    c.Community,
		SignData:     "",
		Lang:         c.Lang})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}