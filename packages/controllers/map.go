package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type mapPage struct {
	Lang             map[string]string
	UserId           int64
}

func (c *Controller) Map() (string, error) {
	TemplateStr, err := makeTemplate("map", "map", &mapPage{
		Lang:             c.Lang,
		UserId:           c.SessUserId,
	})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
