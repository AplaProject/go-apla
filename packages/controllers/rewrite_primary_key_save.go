package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) RewritePrimaryKeySave() (string, error) {

	if !c.NodeAdmin || c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	if len(c.r.FormValue("n")) > 0 {

		c.r.ParseForm()
		n := []byte(c.r.FormValue("n"))
		e := []byte(c.r.FormValue("e"))
		if !utils.CheckInputData(n, "hex") {
			return "", utils.ErrInfo(errors.New("incorrect n"))
		}
		if !utils.CheckInputData(e, "hex") {
			return "", utils.ErrInfo(errors.New("incorrect e"))
		}
		publicKey := utils.MakeAsn1(n, e)

		// проверим, есть ли вообще такой публичный ключ
		userId, err := c.Single("SELECT user_id FROM users WHERE hex(public_key_0) = ?", publicKey).Int64()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if userId == 0 {
			return "", utils.ErrInfo(errors.New("incorrect public_key"))
		}

		// может быть юзер уже майнер?
		minerId, err := c.GetMinerId(userId)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		status := ""
		if minerId > 0 {
			status = "miner"
		} else {
			status = "user"
		}

		err = c.ExecSql(`DELETE FROM my_keys`)
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		err = c.ExecSql(`INSERT INTO `+c.MyPrefix+`my_keys (public_key, status) VALUES ([hex], ?)`, publicKey, "approved")
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		err = c.ExecSql(`UPDATE `+c.MyPrefix+`my_table SET user_id = ?, miner_id = ?, status = ?`, userId, minerId, status)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}
	return `{"success":"success"}`, nil
}
