package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type installStep0Struct struct {
	Lang map[string]string
	//KeyPassword string
}

// Шаг 1 - выбор либо стандартных настроек (sqlite и блокчейн с сервера) либо расширенных - pg/mysql и загрузка с нодов
func (c *Controller) InstallStep0() (string, error) {

	//keyPassword := c.r.FormValue("key_password")

	TemplateStr, err := makeTemplate("install_step_0", "installStep0", &installStep0Struct{
		Lang: c.Lang})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
