package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type adminVariablesPage struct {
	Variables    map[string]string
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

func (c *Controller) AdminVariables() (string, error) {

	txType := "AdminVariables"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	variables, err := c.GetMap(`SELECT * FROM variables`, "name", "value")

	TemplateStr, err := makeTemplate("admin_variables", "adminVariables", &adminVariablesPage{
		Variables:    variables,
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
