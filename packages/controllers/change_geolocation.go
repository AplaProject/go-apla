package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/geolocation"
	"fmt"
	"strings"
)

type changeGeolocationPage struct {
	SignData      string
	ShowSignData  bool
	TxType        string
	TxTypeId      int64
	TimeNow       int64
	UserId        int64
	Alert         string
	Lang          map[string]string
	CountSignArr  []int
	MyGeolocation map[string]string
	MyCountry     int
	Countries     []string
	LimitsText    string
}

func (c *Controller) ChangeGeolocation() (string, error) {

	var err error

	txType := "ChangeGeolocation"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	myGeolocationStr, err := c.Single(`SELECT geolocation FROM ` + c.MyPrefix + `my_table`).String()
	if len(myGeolocationStr) == 0 {
		myGeolocationStr = "39.94887, -75.15005"
		if !utils.Mobile() {
			if coord, err := geolocation.GetLocation(); err == nil {
				myGeolocationStr = fmt.Sprintf("%.6f, %.6f", coord.Latitude, coord.Longitude )
			}
		}
	}
	x := strings.Split(myGeolocationStr, ", ")
	myGeolocation := make(map[string]string)
	myGeolocation["lat"] = x[0]
	myGeolocation["lon"] = x[1]
	myCountry, err := c.Single("SELECT country FROM miners_data WHERE user_id = ?", c.SessUserId).Int()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	limitsText := strings.Replace(c.Lang["geolocation_limits_text"], "[limit]", utils.Int64ToStr(c.Variables.Int64["limit_change_geolocation"]), -1)
	limitsText = strings.Replace(limitsText, "[period]", c.Periods[c.Variables.Int64["limit_change_geolocation_period"]], -1)

	TemplateStr, err := makeTemplate("change_geolocation", "changeGeolocation", &changeGeolocationPage{
		Alert:         c.Alert,
		Lang:          c.Lang,
		TxType:        txType,
		TxTypeId:      txTypeId,
		TimeNow:       timeNow,
		UserId:        c.SessUserId,
		CountSignArr:  c.CountSignArr,
		ShowSignData:  c.ShowSignData,
		SignData:      "",
		LimitsText:    limitsText,
		MyCountry:     myCountry,
		Countries:     consts.Countries,
		MyGeolocation: myGeolocation})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
