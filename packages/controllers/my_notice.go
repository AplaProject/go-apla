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

	myNotice, err := c.GetMyNoticeData(c.SessRestricted, c.SessUserId, c.MyPrefix, c.Lang)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	var cashRequests int64
	if c.SessRestricted == 0 {
		myUserId, err := c.GetMyUserId(c.MyPrefix)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		cashRequests, err = c.Single("SELECT count(id) FROM cash_requests WHERE to_user_id  =  ? AND status  =  'pending' AND for_repaid_del_block_id  =  0 AND del_block_id  =  0", myUserId).Int64()
		if cashRequests > 0 {
			cashRequests = 1
		}
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}

	result, err := json.Marshal(map[string]string{
		"main_status":          myNotice["main_status"],
		"main_status_complete": myNotice["main_status_complete"],
		"account_status":       myNotice["account_status"],
		"cur_block_id":         myNotice["cur_block_id"],
		"connections":          myNotice["connections"],
		"time_last_block":      myNotice["time_last_block"],
		"time_last_block_int":  myNotice["time_last_block_int"],
		"inbox":                utils.Int64ToStr(cashRequests)})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return string(result), nil
}
