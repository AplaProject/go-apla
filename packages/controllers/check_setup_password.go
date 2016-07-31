package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) CheckSetupPassword() (string, error) {

	c.r.ParseForm()
	password, err := c.Single(`SELECT setup_password FROM config WHERE setup_password = ?`, utils.DSha256(c.r.FormValue("password"))).String()
	if err != nil {
		return "", err
	}
	if len(password) > 0 && !c.Community {
		userId, err := c.GetMyUserId("")
		if err != nil {
			return "", err
		}
		publicKey, err := c.GetUserPublicKey(userId)
		if err != nil {
			return "", err
		}
		c.sess.Set("user_id", userId)
		c.sess.Set("public_key", string(utils.BinToHex(publicKey)))
		log.Debug("public_key check: %s", string(utils.BinToHex(publicKey)))
		return "ok", nil
	}
	return "", nil

}
