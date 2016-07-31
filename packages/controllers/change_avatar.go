package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

type ChangeAvatarPage struct {
	SignData        string
	ShowSignData    bool
	TxType          string
	TxTypeId        int64
	TimeNow         int64
	UserId          int64
	Alert           string
	Lang            map[string]string
	CountSignArr    []int
//	LastTxFormatted string
	Avatar          string
	Name            string
}

func (c *Controller) ChangeAvatar() (string, error) {

	txType := "UserAvatar"
	txTypeId := utils.TypeInt(txType)
	timeNow := time.Now().Unix()

/*	last_tx, err := c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"UserAvatar"}), 1, c.TimeFormat)
	lastTxFormatted := ""
	if len(last_tx) > 0 {
		lastTxFormatted, _ = utils.MakeLastTx(last_tx, c.Lang)
	}*/

	data, err := c.OneRow("SELECT name, avatar FROM users WHERE user_id =  ?", c.SessUserId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	avatar := data["avatar"]
	name := data["name"]

	TemplateStr, err := makeTemplate("change_avatar", "changeAvatar", &ChangeAvatarPage{
		Alert:           c.Alert,
		Lang:            c.Lang,
		CountSignArr:    c.CountSignArr,
		ShowSignData:    c.ShowSignData,
		UserId:          c.SessUserId,
		TimeNow:         timeNow,
		TxType:          txType,
		TxTypeId:        txTypeId,
		SignData:        "",
//		LastTxFormatted: lastTxFormatted,
		Avatar:          avatar,
		Name:            name})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
