package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	l "log"
	"os"
	"strings"
	"github.com/DayLightProject/go-daylight/packages/detector"
	"fmt"
)

type upgrade3Page struct {
	Alert           string
	SignData        string
	ShowSignData    bool
	CountSignArr    []int
	UserId          int64
	Lang            map[string]string
	SaveAndGotoStep string
	UpgradeMenu     string
	FaceCoords      string
	ProfileCoords   string
	UserProfile     string
	UserFace        string
	ExamplePoints   map[string]string
	Mobile          bool
	IOS             bool
	Full            bool
}

func (c *Controller) Upgrade3() (string, error) {

	log.Debug("Upgrade3")

	userProfile := *utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_profile.jpg"
	userFace := *utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_face.jpg"

	r, err := detector.Race(userFace)
	if err != nil {
		l.Println(err)
	}

	fmt.Println("race", r)
	l.Println("Race detected:", r)
	err = c.ExecSql("UPDATE "+c.MyPrefix+"my_table SET race = ?", r)

	if err != nil {
		l.Println(err)
	}


	if _, err := os.Stat(userProfile); os.IsNotExist(err) {
		userProfile = ""
	} else {
		userProfile = "public/" + utils.Int64ToStr(c.SessUserId) + "_user_profile.jpg?r=" + utils.IntToStr(utils.RandInt(0, 99999))
	}
	if _, err := os.Stat(userFace); os.IsNotExist(err) {
		userFace = ""
	} else {
		userFace = "public/" + utils.Int64ToStr(c.SessUserId) + "_user_face.jpg?r=" + utils.IntToStr(utils.RandInt(0, 99999))
	}

	log.Debug("userProfile: %s", userProfile)
	log.Debug("userFace: %s", userFace)

	l.Printf("userProfile: %s", userProfile)
	l.Printf("userFace: %s", userFace)

	// текущий набор точек для шаблонов
	// current set of points for the templates
	examplePoints, err := c.GetPoints(c.Lang)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	// точки, которые юзер уже отмечал
	// user selected points
	data, err := c.OneRow("SELECT face_coords, profile_coords FROM " + c.MyPrefix + "my_table").String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	faceCoords := ""
	profileCoords := ""
	if len(data["face_coords"]) > 0 {
		faceCoords = data["face_coords"]
		profileCoords = data["profile_coords"]
	}
	upgradeMenu, full, next := utils.MakeUpgradeMenu(2)
	saveAndGotoStep := strings.Replace(c.Lang["save_and_goto_step"], "[num]", next, -1)

	TemplateStr, err := makeTemplate("upgrade_3", "upgrade3", &upgrade3Page{
		Alert:           c.Alert,
		Lang:            c.Lang,
		CountSignArr:    c.CountSignArr,
		ShowSignData:    c.ShowSignData,
		SaveAndGotoStep: saveAndGotoStep,
		UpgradeMenu:     upgradeMenu,
		UserId:          c.SessUserId,
		FaceCoords:      faceCoords,
		ProfileCoords:   profileCoords,
		UserProfile:     userProfile,
		UserFace:        userFace,
		Full:            full,
		ExamplePoints:   examplePoints,
		IOS:             utils.IOS(),
		Mobile:          utils.Mobile()})
	if err != nil {
		return "", utils.ErrInfo(err)
	}


	return TemplateStr, nil
}
