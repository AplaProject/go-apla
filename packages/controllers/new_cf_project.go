package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type NewCfProjectPage struct {
	Alert        string
	SignData     string
	ShowSignData bool
	Lang         map[string]string
	UserId       int64
	TxType       string
	TxTypeId     int64
	TimeNow      int64
	CountSignArr []int
	CurrencyList map[int64]string
	Latitude     string
	Longitude    string
	City         string
	EndTime      int64
	CountDaysArr []int
	CfCategory   []map[string]string
}

func (c *Controller) NewCfProject() (string, error) {

	var err error

	txType := "NewCfProject"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	cfCategory := utils.MakeCfCategories(c.Lang)

	latitude := "39.94887"
	longitude := "-75.15005"
	city := "Pennsylvania, USA"
	endTime := utils.Time() + 3600*24*7 + 3600

	var countDaysArr []int
	for i := 7; i < 90; i++ {
		countDaysArr = append(countDaysArr, i)
	}

	TemplateStr, err := makeTemplate("new_cf_project", "newCfProject", &NewCfProjectPage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		CountSignArr: c.CountSignArr,
		ShowSignData: c.ShowSignData,
		UserId:       c.SessUserId,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId,
		Latitude:     latitude,
		Longitude:    longitude,
		EndTime:      endTime,
		City:         city,
		CfCategory:   cfCategory,
		CountDaysArr: countDaysArr,
		CurrencyList: c.CurrencyList})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
