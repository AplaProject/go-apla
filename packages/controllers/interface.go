package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type InterfacePage struct {
	Show_sign_data          int64
	Show_map                int64
	Show_progress_bar       int64
	Alert                   string
	Param_show_progress_bar int64
	Lang                    map[string]string
}

func (c *Controller) Interface() (string, error) {

	if c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	var err error
	name := ""
	if len(c.Parameters["show_map"]) > 0 {
		name = "show_map"
	} else if len(c.Parameters["show_sign_data"]) > 0 {
		name = "show_sign_data"
	} else if len(c.Parameters["show_progress_bar"]) > 0 {
		name = "show_progress_bar"
	}
	alert := ""
	if len(name) > 0 {
		err = c.ExecSql("UPDATE "+c.MyPrefix+"my_table SET "+name+" = ?", utils.StrToInt64(c.Parameters[name]))
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		alert = c.Lang["done"]
	}

	data, err := c.OneRow("SELECT * FROM " + c.MyPrefix + "my_table").Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	show_sign_data := data["show_sign_data"]
	show_map := data["show_map"]
	show_progress_bar := data["show_progress_bar"]

	param_show_progress_bar := utils.StrToInt64(c.Parameters["show_progress_bar"])

	TemplateStr, err := makeTemplate("interface", "interface", &InterfacePage{
		Lang:                    c.Lang,
		Show_sign_data:          show_sign_data,
		Show_map:                show_map,
		Show_progress_bar:       show_progress_bar,
		Param_show_progress_bar: param_show_progress_bar,
		Alert: alert})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
