package controllers

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) EGetBalance() (string, error) {

	var myWallets []map[string]string
	myWallets, err := c.getMyWallets()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	html := ""
	for _, data := range myWallets {
		html += fmt.Sprintf("<strong>%v</strong> %v<br>", data["amount"], data["currency_name"])
	}

	return utils.JsonAnswer(html, "html").String(), nil
}
