package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"regexp"
)

func (c *Controller) SaveVideo() (string, error) {

	if c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	c.r.ParseForm()
	videoUrl := c.r.FormValue("video_url")
	videoType := ""
	videoId := ""
	re, _ := regexp.Compile(`(?i)youtu\.be\/([\w\-]+)`)
	match := re.FindStringSubmatch(videoUrl)
	if len(match) > 0 {
		videoType = "youtube"
		videoId = match[1]
	} else {
		re, _ := regexp.Compile(`(?i)embed\/([\w\-]+)`)
		match := re.FindStringSubmatch(videoUrl)
		if len(match) > 0 {
			videoType = "youtube"
			videoId = match[1]
		} else {
			re, _ := regexp.Compile(`(?i)watch\?v=([\w\-]+)`)
			match := re.FindStringSubmatch(videoUrl)
			if len(match) > 0 {
				videoType = "youtube"
				videoId = match[1]
			}
		}
	}
	if len(videoType) > 0 {
		err := c.ExecSql("UPDATE "+c.MyPrefix+"my_table SET video_url_id = ?, video_type = ?", videoId, videoType)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		return `{"url":"http://www.youtube.com/embed/` + videoId + `"}`, nil
	} else {
		return `{"url":""}`, nil
	}
	return ``, nil
}
