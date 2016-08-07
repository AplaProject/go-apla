package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"strings"
)

func (c *Controller) Update() (string, error) {

	ver, _, err := utils.GetUpdVerAndUrl(consts.UPD_AND_VER_URL)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if len(ver) > 0 {
		newVersion := strings.Replace(c.Lang["new_version"], "[ver]", ver, -1)
		return utils.JsonAnswer(newVersion, "success").String(), nil
	}
	return "", nil
}
