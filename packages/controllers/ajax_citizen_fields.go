package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

const ACitizenFields = `ajax_citizen_fields`

type CitizenFieldsJson struct {
	Data   string    `json:"data"`
	Error  string    `json:"error"`
}

func init() {
	newPage(ACitizenFields, `json`)
}

func (c *Controller) AjaxCitizenFields() interface{} {
	var result CitizenFieldsJson
	data,err := c.Single(`SELECT fields FROM citizen_fields where state_id=?`, 
					utils.StrToInt(c.r.FormValue("state_id"))).String()
	if err != nil {
		result.Error = err.Error()
	} else {
		result.Data = data
	}
	return result
}
