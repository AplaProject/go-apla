package controllers

import (
	"fmt"
)

const NMenu = `menu`

type menuPage struct {
	Data *CommonPage
}

func init() {
	newPage(NMenu)
}

func (c *Controller) Menu() (string, error) {
	fmt.Println(`Menu Page`)
	return proceedTemplate(c, NMenu, &menuPage{Data: c.Data})
}
