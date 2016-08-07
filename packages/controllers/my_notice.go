package controllers

import (
	"encoding/json"
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) MyNoticeData() (string, error) {

	if !c.dbInit {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	c.r.ParseForm()

	myNotice, err := c.GetMyNoticeData(c.SessCitizenId, c.SessWalletId, c.Lang)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	result, err := json.Marshal(map[string]string{
		"main_status":          myNotice["main_status"],
		"main_status_complete": myNotice["main_status_complete"],
		"account_status":       myNotice["account_status"],
		"cur_block_id":         myNotice["cur_block_id"],
		"connections":          myNotice["connections"],
		"time_last_block":      myNotice["time_last_block"],
		"time_last_block_int":  myNotice["time_last_block_int"]})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return string(result), nil
}
