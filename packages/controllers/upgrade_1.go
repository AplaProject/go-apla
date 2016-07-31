package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"

	"os"
	"strings"
)

type upgrade1Page struct {
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

func (c *Controller) Upgrade1() (string, error) {

	log.Debug("Upgrade1")

	userFace := ""
	/*userProfile := ""

	path := "public/"+utils.Int64ToStr(c.SessUserId)+"_user_profile.jpg"
	if _, err := os.Stat(path); err == nil {
		userProfile = path
	}*/

	path := *utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_face.jpg"
	if _, err := os.Stat(path); err == nil {
		userFace = "public/" + utils.Int64ToStr(c.SessUserId) + "_user_face.jpg"
	}

	step := "1"
	photoType := "face"
	photo := userFace

	upgradeMenu, full, next := utils.MakeUpgradeMenu(0)
	saveAndGotoStep := strings.Replace(c.Lang["save_and_goto_step"], "[num]", next, -1)
	nextStep := "2"
	
	if !full {
		nextStep = "3"
	}
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
