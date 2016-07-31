package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type setupPasswordPage struct {
	Lang map[string]string
}

func (c *Controller) SetupPassword() (string, error) {

	TemplateStr, err := makeTemplate("setup_password", "setupPassword", &setupPasswordPage{
		Lang: c.Lang})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
