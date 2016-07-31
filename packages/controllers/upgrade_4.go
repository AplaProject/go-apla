package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"os"
	"strings"
)

type upgrade4Page struct {
	Alert           string
	UserId          int64
	Lang            map[string]string
	VideoUrl        string
	SaveAndGotoStep string
	UpgradeMenu     string
	UserVideoMp4    string
	UserVideoWebm   string
	UserVideoOgg    string
	Mobile          bool
}

func (c *Controller) Upgrade4() (string, error) {

	log.Debug("Upgrade4")

	videoUrl := ""

	// есть ли загруженное видео.
	data, err := c.OneRow("SELECT video_url_id, video_type FROM " + c.MyPrefix + "my_table").String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	switch data["video_type"] {
	case "youtube":
		videoUrl = "http://www.youtube.com/embed/" + data["video_url_id"]
	case "vimeo":
		videoUrl = "http://www.vimeo.com/embed/" + data["video_url_id"]
	case "youku":
		videoUrl = "http://www.youku.com/embed/" + data["video_url_id"]
	}

	upgradeMenu,_,next := utils.MakeUpgradeMenu(3)
	saveAndGotoStep := strings.Replace(c.Lang["save_and_goto_step"], "[num]", next, -1)

	var userVideoMp4 string
	path := *utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_video.mp4"
	if _, err := os.Stat(path); err == nil {
		userVideoMp4 = "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_video.mp4"
	}
	var userVideoWebm string
	path = *utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_video.webm"
	if _, err := os.Stat(path); err == nil {
		userVideoWebm = "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_video.webm"
	}
	var userVideoOgg string
	path = *utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_video.ogg"
	if _, err := os.Stat(path); err == nil {
		userVideoOgg = "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_video.ogg"
	}

	TemplateStr, err := makeTemplate("upgrade_4", "upgrade4", &upgrade4Page{
		Alert:           c.Alert,
		Lang:            c.Lang,
		SaveAndGotoStep: saveAndGotoStep,
		UpgradeMenu:     upgradeMenu,
		VideoUrl:        videoUrl,
		UserVideoMp4:    userVideoMp4,
		UserVideoWebm:   userVideoWebm,
		UserVideoOgg:    userVideoOgg,
		Mobile:          utils.Mobile(),
		UserId:          c.SessUserId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
