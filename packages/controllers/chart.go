package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type chartPage struct {
	Chart					string
	DCTarget int64
}

func (c *Controller) Chart() (string, error) {

	var chart string
	// график обещанные суммы/монеты
	chartData, err := c.GetAll(`
			SELECT month, day, dc, promised_amount
			FROM stats
			WHERE currency_id = 72
			ORDER BY id DESC
			LIMIT 7`, 7)
	for i:=len(chartData)-1; i>=0; i-- {
		chart += `['`+chartData[i]["month"]+`/`+chartData[i]["day"]+`', `+utils.ClearNull(chartData[i]["promised_amount"], 0)+`, `+utils.ClearNull(chartData[i]["dc"], 0)+`],`
	}
	if len(chart) > 0 {
		chart = chart[:len(chart)-1]
	}

	DCTarget := consts.DCTarget[72]

	TemplateStr, err := makeTemplate("chart", "chart", &chartPage{
		DCTarget: DCTarget,
		Chart: 					chart})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}