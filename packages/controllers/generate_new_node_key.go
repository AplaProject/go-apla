package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"encoding/json"
)

func (c *Controller) GenerateNewNodeKey() (string, error) {

	priv, pub := utils.GenKeys()
	json, err := json.Marshal(map[string]string{"private_key": priv, "public_key": pub,
								"time": utils.Int64ToStr(utils.Time()) })
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	log.Debug("%v", json)
	return string(json), nil
}
