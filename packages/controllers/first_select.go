package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type firstSelectPage struct {
	Lang map[string]string
}

func (c *Controller) FirstSelect() (string, error) {

	TemplateStr, err := makeTemplate("first_select", "firstSelect", &firstSelectPage{
		Lang: c.Lang})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
