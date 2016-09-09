package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

const NRequestCitizen = `request_citizen_status`

type citizenPage struct {
	Data       *CommonPage
	TxType       string
	TxTypeId     int64
}

func init() {
	newPage(NRequestCitizen)
}

func (c *Controller) RequestCitizenStatus() (string, error) {
	txType := "CitizenRequest"
	pageData := citizenPage{Data:c.Data, TxType: txType, TxTypeId: utils.TypeInt(txType)}
	return proceedTemplate( c, NRequestCitizen, &pageData )
}
