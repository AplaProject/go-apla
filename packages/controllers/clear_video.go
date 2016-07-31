package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) ClearVideo() (string, error) {

	if c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	err := c.ExecSql("UPDATE "+c.MyPrefix+"my_table SET video_url_id = ?, video_type = ?", "", "")
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	return ``, nil
}
