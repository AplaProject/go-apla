package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"strings"
)

type upgrade0Page struct {
	Alert           string
	SignData        string
	ShowSignData    bool
	CountSignArr    []int
	UserId          int64
	Lang            map[string]string
	SaveAndGotoStep string
	UpgradeMenu     string
	Countries       []string
	Country         int
	Race            int
	Mobile          bool
}


func (c *Controller) Upgrade0() (string, error) {

	log.Debug("first_select: %v", c.Parameters["first_select"])
	if c.Parameters["first_select"] == "1" {
		c.ExecSql(`UPDATE ` + c.MyPrefix + `my_table SET first_select=1`)
	}

	data, err := c.OneRow("SELECT race, country FROM " + c.MyPrefix + "my_table").Int()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	race := data["race"]
	country := 0
	if race > 0 {
		country = data["country"]
	}

	saveAndGotoStep := strings.Replace(c.Lang["save_and_goto_step"], "[num]", "2", -1)
	upgradeMenu,_,_ := utils.MakeUpgradeMenu(0)

	TemplateStr, err := makeTemplate("upgrade_0", "upgrade0", &upgrade0Page{
		Alert:           c.Alert,
		Lang:            c.Lang,
		CountSignArr:    c.CountSignArr,
		ShowSignData:    c.ShowSignData,
		SaveAndGotoStep: saveAndGotoStep,
		UpgradeMenu:     upgradeMenu,
		UserId:          c.SessUserId,
		Countries:       consts.Countries,
		Country:         country,
		Race:            race,
		Mobile:          utils.Mobile()})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
