package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

func (c *Controller) Profile() (string, error) {

	c.r.ParseForm()

	userId := int64(utils.StrToFloat64(c.r.FormValue("user_id")))

	// получаем кол-во TDC на обещанных суммах
	rows, err := c.Query(c.FormatQuery(`
			SELECT from_user_id, time, comment
			FROM abuses
			WHERE user_id = ?
			`), userId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	abuses := ""
	defer rows.Close()
	for rows.Next() {
		var from_user_id, abusestime int64
		var comment string
		err = rows.Scan(&from_user_id, &abusestime, &comment)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		t := time.Unix(abusestime, 0)
		abuses += fmt.Sprintf("from_user_id: %d; time: %s; comment: %s<br>", from_user_id, t.Format(c.TimeFormat), comment)
	}
	if len(abuses) == 0 {
		abuses = "No"
	}
	regTime, err := c.Single("SELECT reg_time FROM miners_data WHERE user_id  =  ?", userId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	t := time.Unix(regTime, 0)
	result, err := json.Marshal(map[string]string{"abuses": abuses, "reg_time": t.Format(c.TimeFormat)})
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	return string(result), nil
}
