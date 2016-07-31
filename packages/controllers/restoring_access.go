package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

type restoringAccessPage struct {
	Alert           string
	SignData        string
	ShowSignData    bool
	CountSignArr    []int
	Lang            map[string]string
	UserId          int64
	TxType          string
	TxTypeId        int64
	TimeNow         int64
	AdminUserId     int64
	ChangeKeyStatus string
	Requests        string
}

func (c *Controller) RestoringAccess() (string, error) {

	var err error

	txType := "ChangeKeyActive"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	adminUserId, err := c.GetAdminUserId()

	data, err := c.OneRow("SELECT change_key, change_key_time, change_key_close FROM users WHERE user_id  =  ?", c.SessUserId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	// разрешил ли уже юзер менять свой ключ админу
	changeKeyStatus := data["change_key"]

	requests := ""
	if len(data["change_key_status"]) > 0 && len(data["change_key_close"]) > 0 {
		t := time.Unix(utils.StrToInt64(data["change_key_time"]), 0)
		requests = t.Format(c.TimeFormat)
	}

	TemplateStr, err := makeTemplate("restoring_access", "restoringAccess", &restoringAccessPage{
		Alert:           c.Alert,
		Lang:            c.Lang,
		ShowSignData:    c.ShowSignData,
		SignData:        "",
		UserId:          c.SessUserId,
		CountSignArr:    c.CountSignArr,
		TimeNow:         timeNow,
		TxType:          txType,
		TxTypeId:        txTypeId,
		AdminUserId:     adminUserId,
		ChangeKeyStatus: changeKeyStatus,
		Requests:        requests})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
