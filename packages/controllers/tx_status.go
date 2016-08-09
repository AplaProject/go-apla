package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) TxStatus() (string, error) {

	hash := c.r.FormValue("hash")

	tx, err := c.OneRow(`SELECT block_id, error FROM transactions_status WHERE hash = [hex]`, hash).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if tx["block_id"] != "0" && tx["block_id"] != "" {
		return `{"success":"`+tx["block_id"]+`"}`, nil
	} else if len(tx["error"]) > 0 {
		return `{"error":"`+tx["block_id"]+`"}`, nil
	}
	return "", nil
}