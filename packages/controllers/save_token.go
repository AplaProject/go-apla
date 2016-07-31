package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) SaveToken() (string, error) {

	if c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	c.r.ParseForm()
	token := c.r.FormValue("token")
	eOwnerId := c.r.FormValue("e_owner_id")

	if len(token) > 0 {
		exists, err := c.Single(`SELECT token FROM `+c.MyPrefix+`my_tokens WHERE token = ?`, token).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if len(exists) == 0 {
			err := c.ExecSql(`INSERT INTO `+c.MyPrefix+`my_tokens (token, e_owner_id, time) VALUES (?, ?, ?)`, token, eOwnerId, utils.Time())
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		}
	}
	return ``, nil
}
