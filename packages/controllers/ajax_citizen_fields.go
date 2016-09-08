package controllers

import (
//	"github.com/DayLightProject/go-daylight/packages/utils"
)

const ACitizenFields = `ajax_citizen_fields`

type CitizenFieldsJson struct {
	Fields  string `json:"fields"`
	Price   int64  `json:"price"`
	Valid   bool   `json:"valid"`
	Error   string `json:"error"`
}

func init() {
	newPage(ACitizenFields, `json`)
}

func (c *Controller) AjaxCitizenFields() interface{} {
	var result CitizenFieldsJson
	var err error
	result.Fields,err = c.Single(`SELECT value FROM ds_state_settings where parameter='citizen_fields'`).String()
	if err == nil {
		result.Price, err = c.Single(`SELECT value FROM ds_state_settings where parameter='citizen_dlt_price'`).Int64()
		if err == nil {
			amount,err := c.Single("select amount from dlt_wallets where wallet_id=?", c.SessWalletId ).Int64()
			result.Valid = (err == nil && amount >= result.Price)
		}
	}
	if err != nil {
		result.Error = err.Error()
	}
	return result
}
