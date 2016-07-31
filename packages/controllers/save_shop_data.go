package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) SaveShopData() (string, error) {

	if c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	c.r.ParseForm()
	shop_secret_key := c.r.FormValue("shop_secret_key")
	shop_callback_url := c.r.FormValue("shop_callback_url")
	err := c.ExecSql("UPDATE "+c.MyPrefix+"my_table SET shop_secret_key = ?, shop_callback_url = ?", shop_secret_key, shop_callback_url)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return "ok", nil
}
