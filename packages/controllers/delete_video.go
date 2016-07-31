package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"os"
)

func (c *Controller) DeleteVideo() (string, error) {

	if !c.NodeAdmin || c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	if c.r.FormValue("type") == "mp4" {
		err := os.Remove(*utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_video.mp4")
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	} else if c.r.FormValue("type") == "webm_ogg" {
		err := os.Remove(*utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_video.ogv")
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		err = os.Remove(*utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_video.webm")
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}

	return ``, nil
}
