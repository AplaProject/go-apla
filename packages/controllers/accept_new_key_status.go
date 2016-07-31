package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) AcceptNewKeyStatus() (string, error) {

	if c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	result := ""
	status, err := c.DCDB.Single("SELECT status FROM " + c.MyPrefix + "my_table").String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	log.Debug(">status: ", status)
	if status == "user" {
		result = "ok"
	}

	return result, nil
}
