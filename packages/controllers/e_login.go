package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) ELogin() (string, error) {

	c.r.ParseForm()
	email := c.r.FormValue("email")
	password := c.r.FormValue("password")
	if len(password) > 50 || len(password) < 1 {
		return "", errors.New(c.Lang["invalid_pass"])
	}

	data, err := c.OneRow("SELECT id, salt FROM e_users WHERE email = ?", email).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if len(data) == 0 {
		return "", errors.New(c.Lang["email_is_not_registered"])
	}

	// проверяем, верный ли пароль
	passAndSalt := utils.Sha256(password + data["salt"])
	userId, err := utils.DB.Single("SELECT id FROM e_users WHERE id  =  ? AND password  =  ?", data["id"], passAndSalt).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if userId == 0 {
		return "", errors.New(c.Lang["wrong_pass"])
	}

	c.sess.Set("e_user_id", userId)

	return utils.JsonAnswer("success", "success").String(), nil
}
