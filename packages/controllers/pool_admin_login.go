package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type poolAdminPageLogin struct {
}

func (c *Controller) PoolAdminLogin() (string, error) {

	if !c.PoolAdmin {
		return "", utils.ErrInfo(errors.New("access denied"))
	}

	c.r.ParseForm()
	userId := int64(utils.StrToFloat64(c.Parameters["userId"]))
	publicKey, err := c.GetUserPublicKey(userId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	c.sess.Set("user_id", userId)
	c.sess.Set("public_key", string(utils.BinToHex(publicKey)))

	TemplateStr, err := makeTemplate("pool_admin_login", "poolAdminLogin", &poolAdminPage{})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
