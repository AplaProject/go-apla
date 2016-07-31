package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) SaveEmailSms() (string, error) {

	if !c.NodeAdmin || c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	c.r.ParseForm()

	err := c.ExecSql(`
			UPDATE `+c.MyPrefix+`my_table
			SET  email = ?,
					smtp_server =  ?,
					use_smtp =  ?,
					smtp_port =  ?,
					smtp_ssl =  ?,
					smtp_auth =  ?,
					smtp_username = ?,
					smtp_password = ?,
					sms_http_get_request = ?
			`, c.r.FormValue("email"), c.r.FormValue("smtp_server"), c.r.FormValue("use_smtp"), c.r.FormValue("smtp_port"), c.r.FormValue("smtp_ssl"), c.r.FormValue("smtp_auth"), c.r.FormValue("smtp_username"), c.r.FormValue("smtp_password"), c.r.FormValue("sms_http_get_request"))
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	return "", nil

}
