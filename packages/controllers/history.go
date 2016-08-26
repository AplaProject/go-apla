package controllers

const NHistory = `history`

type historyPage struct {
	Data       *CommonPage
}

func init() {
	newPage(NHistory)
}

func (c *Controller) History() (string, error) {
	pageData := historyPage{Data:c.Data}
	return proceedTemplate( c, NHistory, &pageData )
}
