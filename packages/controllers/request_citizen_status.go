package controllers

const NRequestCitizen = `request_citizen_status`

type citizenPage struct {
	Data       *CommonPage
}

func init() {
	newPage(NRequestCitizen)
}

func (c *Controller) RequestCitizenStatus() (string, error) {
	pageData := citizenPage{Data:c.Data}
	return proceedTemplate( c, NRequestCitizen, &pageData )
}
