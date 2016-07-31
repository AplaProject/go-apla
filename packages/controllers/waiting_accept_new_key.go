package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type waitingAcceptNewKeyPage struct {
	Lang map[string]string
}

func (c *Controller) WaitingAcceptNewKey() (string, error) {

	if c.SessUserId > 0 {
		err := c.SendTxChangePkey(c.SessUserId)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		err = c.ExecSql(`UPDATE ` + c.MyPrefix + `my_table SET status='waiting_accept_new_key'`)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}

	TemplateStr, err := makeTemplate("waiting_accept_new_key", "waitingAcceptNewKey", &waitingAcceptNewKeyPage{
		Lang: c.Lang})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
