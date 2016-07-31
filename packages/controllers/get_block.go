package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	//	"encoding/json"
)

func (c *Controller) GetBlock() (string, error) {

	c.r.ParseForm()

	BlockId := int64(utils.StrToFloat64(c.r.FormValue("id")))
	if BlockId == 0 {
		return `{"error": "nil id"}`, nil
	}

	if len(c.r.FormValue("download")) > 0 {
		c.w.Header().Set("Content-type", "application/octet-stream")
		c.w.Header().Set("Content-Disposition", "attachment; filename=\""+utils.Int64ToStr(BlockId)+".binary\"")
	}

	block, err := c.Single("SELECT data FROM block_chain WHERE id  =  ?", BlockId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return block, nil
}
