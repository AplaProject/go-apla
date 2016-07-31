package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type ChangeCountryRacePage struct {
	Alert     string
	Lang      map[string]string
	Countries []string
	Country   int
	Race      int
}

func (c *Controller) ChangeCountryRace() (string, error) {

	data, err := c.OneRow("SELECT race, country FROM " + c.MyPrefix + "my_table").Int()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	TemplateStr, err := makeTemplate("change_country_race", "changeCountryRace", &ChangeCountryRacePage{
		Alert:     c.Alert,
		Lang:      c.Lang,
		Countries: consts.Countries,
		Country:   data["country"],
		Race:      data["race"]})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
