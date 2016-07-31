package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) ESignUp() (string, error) {

	c.r.ParseForm()
	email := c.r.FormValue("email")
	password := c.r.FormValue("password")
	if len(password) > 50 || len(password) < 1 {
		return "", errors.New(c.Lang["invalid_pass"])
	}

	existsEmail, err := c.Single("SELECT id FROM e_users WHERE email = ?", email).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if existsEmail > 0 {
		return "", errors.New(c.Lang["exists_email"])
	}

	salt := utils.RandSeq(32)
	passAndSalt := utils.DSha256(password + salt)
	userId, err := c.ExecSqlGetLastInsertId("INSERT INTO e_users ( email, password, ip, salt ) VALUES ( ?, ?, ?, ? )", "id", email, passAndSalt, c.r.RemoteAddr, salt)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	c.sess.Set("e_user_id", userId)

	return utils.JsonAnswer("success", "success").String(), nil
}
