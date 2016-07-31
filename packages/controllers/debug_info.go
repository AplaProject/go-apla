package controllers

import (
	"encoding/json"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) DebugInfo() (string, error) {

	mainLock, err := c.OneRow(`SELECT * FROM main_lock`).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	jsonMainLock, err := json.Marshal(mainLock)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	nodesBan, err := c.GetAll(`SELECT * FROM nodes_ban`, 20)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	jsonNodesBan, err := json.Marshal(nodesBan)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	return consts.VERSION + "\n" + string(jsonMainLock) + "\n" + string(jsonNodesBan) + "\n", nil
}
