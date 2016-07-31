package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type changePoolPage struct {
	Alert        string
	SignData     string
	ShowSignData bool
	CountSignArr []int
	Data         map[string]string
	Lang         map[string]string
	UserId       int64
	LimitsText   string
	Community    bool
	TxType       string
	TxTypeId     int64
	TimeNow      int64
}

func (c *Controller) ChangePool() (string, error) {

	txType := "ChangePool"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	if !c.PoolAdmin || c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("access denied"))
	}

	TemplateStr, err := makeTemplate("change_pool", "changePool", &changePoolPage{
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
