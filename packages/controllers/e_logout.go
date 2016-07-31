package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) ELogout() (string, error) {

	if c.SessUserId == 0 {
		return utils.JsonAnswer("error", "empty SessUserId").String(), nil
	}
	c.sess.Delete("e_user_id")

	return utils.JsonAnswer("success", "success").String(), nil
}
