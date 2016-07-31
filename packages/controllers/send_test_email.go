package controllers

import (
	"errors"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	//"log"
)

func (c *Controller) SendTestEmail() (string, error) {

	if c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	c.r.ParseForm()

	mailData, err := c.OneRow("SELECT * FROM " + c.MyPrefix + "my_table").String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
//	err = c.SendMail("Test", "Test", mailData["email"], mailData, c.Community, c.PoolAdminUserId)
	if err = utils.SendEmail( mailData["email"], utils.StrToInt64(mailData["user_id"]), utils.ECMD_TEST, nil ); err != nil {
		return fmt.Sprintf(`{"error":"%s"}`, err), nil
	}

	return `{"error":"null"}`, nil
}
