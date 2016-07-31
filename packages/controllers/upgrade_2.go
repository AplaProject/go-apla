package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"

	"os"
	"strings"
)

type upgrade2Page struct {
	Alert           string
	SignData        string
	ShowSignData    bool
	CountSignArr    []int
	UserId          int64
	Lang            map[string]string
	SaveAndGotoStep string
	UpgradeMenu     string
	Step            string
	NextStep        string
	PhotoType       string
	Photo           string
	Mobile          bool
	IOS             bool
	Full            bool
}

func (c *Controller) Upgrade2() (string, error) {

	log.Debug("Upgrade2")

	userProfile := ""
	path := *utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_profile.jpg"
	if _, err := os.Stat(path); err == nil {
		userProfile = "public/" + utils.Int64ToStr(c.SessUserId) + "_user_profile.jpg"
	}

	step := "2"
	nextStep := "3"
	photoType := "profile"
	photo := userProfile

	saveAndGotoStep := strings.Replace(c.Lang["save_and_goto_step"], "[num]", "3", -1)
	upgradeMenu, full, _ := utils.MakeUpgradeMenu(1)

	TemplateStr, err := makeTemplate("upgrade_1_and_2", "upgrade1And2", &upgrade1Page{
		Alert:           c.Alert,
		Lang:            c.Lang,
		CountSignArr:    c.CountSignArr,
		ShowSignData:    c.ShowSignData,
		SaveAndGotoStep: saveAndGotoStep,
		UpgradeMenu:     upgradeMenu,
		UserId:          c.SessUserId,
		PhotoType:       photoType,
		Photo:           photo,
		Step:            step,
		Full:            full,
		NextStep:        nextStep,
		IOS:             utils.IOS(),
		Mobile:          utils.Mobile()})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
