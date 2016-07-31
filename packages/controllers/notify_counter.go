// notify_counter
package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"encoding/json"
)

func (c *Controller) NotifyCounter() (string, error) {

	resval := false
	result := func(msg, data string, success bool ) (string, error) {
		res, err := json.Marshal( answerJson{Result:resval, Error: msg,
		                          Data: data, Success: success})
		return string(res), err
	}
	
	userId := utils.StrToInt64( c.r.FormValue(`user_id`))
	
	count, err := c.GetNotificationsCount(userId)
	if err != nil {
		result(err.Error(), ``, false)
	}
	return result( ``, utils.IntToStr( int( count )), true )
}
