package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type AbusePage struct {
	Alert          string
	SignData       string
	ShowSignData   bool
	CountAbusesArr []int
	CountSignArr   []int
	Lang           map[string]string
	UserId         int64
	TxType         string
	TxTypeId       int64
	TimeNow        int64
}

func (c *Controller) Abuse() (string, error) {

	var err error

	txType := "Abuses"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	var countAbusesArr []int
	for i := 0; i < 20; i++ {
		countAbusesArr = append(countAbusesArr, i)
	}

	TemplateStr, err := makeTemplate("abuse", "abuse", &AbusePage{
		Alert:          c.Alert,
		Lang:           c.Lang,
		CountAbusesArr: countAbusesArr,
		ShowSignData:   c.ShowSignData,
		SignData:       "",
		UserId:         c.SessUserId,
		CountSignArr:   c.CountSignArr,
		TimeNow:        timeNow,
		TxType:         txType,
		TxTypeId:       txTypeId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
