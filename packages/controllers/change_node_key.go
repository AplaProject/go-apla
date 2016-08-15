package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type changeNodeKeyPage struct {
	Alert        string
	SignData     string
	ShowSignData bool
	CountSignArr []int
	Lang         map[string]string
	WalletId int64
	CitizenId int64
	TxType       string
	TxTypeId     int64
	TimeNow      int64
}

func (c *Controller) ChangeNodeKey() (string, error) {

	var err error

	txType := "ChangeNodeKey"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	TemplateStr, err := makeTemplate("change_node_key", "changeNodeKey", &changeNodeKeyPage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		ShowSignData: c.ShowSignData,
		SignData:     "",
		WalletId: c.SessWalletId,
		CitizenId: c.SessCitizenId,
		CountSignArr: c.CountSignArr,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
