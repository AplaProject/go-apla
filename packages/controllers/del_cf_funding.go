package controllers

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type DelCfFundingPage struct {
	Alert               string
	SignData            string
	ShowSignData        bool
	Lang                map[string]string
	UserId              int64
	TxType              string
	TxTypeId            int64
	TimeNow             int64
	CountSignArr        []int
	DelId               int64
	ProjectCurrencyName string
}

func (c *Controller) DelCfFunding() (string, error) {

	var err error

	txType := "DelCfFunding"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	delId := int64(utils.StrToFloat64(c.Parameters["del_id"]))

	TemplateStr, err := makeTemplate("del_cf_funding", "delCfFunding", &DelCfFundingPage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		CountSignArr: c.CountSignArr,
		ShowSignData: c.ShowSignData,
		SignData:     fmt.Sprintf(`%d,%d,%d,%d`, txTypeId, timeNow, c.SessUserId, delId),
		UserId:       c.SessUserId,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId,
		DelId:        delId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
