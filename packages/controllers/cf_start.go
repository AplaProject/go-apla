package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type cfStartPage struct {
	Lang   map[string]string
	CfUrl  string
	UserId int64
}

func (c *Controller) CfStart() (string, error) {

	var err error

	cfUrl, err := c.GetCfUrl()

	TemplateStr, err := makeTemplate("cf_start", "cfStart", &cfStartPage{
		Lang:   c.Lang,
		CfUrl:  cfUrl,
		UserId: c.SessUserId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
