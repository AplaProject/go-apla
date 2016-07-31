package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"

	"encoding/json"
	"errors"
)

func (c *Controller) GenerateNewNodeKey() (string, error) {

	if c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	priv, pub := utils.GenKeys()
	json, err := json.Marshal(map[string]string{"private_key": priv, "public_key": pub})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	log.Debug("%v", json)
	return string(json), nil
}
