package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type setPasswordPage struct {
	Lang map[string]string
	IOS  bool
	Android  bool
	Mobile bool
}

func (c *Controller) SetPassword() (string, error) {

	TemplateStr, err := makeTemplate("set_password", "setPassword", &setPasswordPage{
		Lang: c.Lang, IOS: utils.IOS(), Android: utils.Android(), Mobile: utils.Mobile()})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
