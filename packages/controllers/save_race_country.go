package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) SaveRaceCountry() (string, error) {

	if c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	c.r.ParseForm()
	race := int(utils.StrToFloat64(c.r.FormValue("race")))
	country := int(utils.StrToFloat64(c.r.FormValue("country")))
	err := c.ExecSql("UPDATE "+c.MyPrefix+"my_table SET race = ?, country = ?", race, country)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return `{"error":"0"}`, nil
}
