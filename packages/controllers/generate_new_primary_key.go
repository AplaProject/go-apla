package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"

	"encoding/base64"
	"encoding/json"
)

func (c *Controller) GenerateNewPrimaryKey() (string, error) {

	c.r.ParseForm()
	password := c.r.FormValue("password")

	priv, pub := utils.GenKeys()
	if len(password) > 0 {
		log.Debug("priv:", priv)
		encKey, err := utils.Encrypt(utils.Md5("11"), []byte(priv))
		log.Debug("priv encKey:", encKey)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		priv = base64.StdEncoding.EncodeToString(encKey)
		log.Debug("priv ENC:", priv)
		//priv = string(encKey)
	}
	json, err := json.Marshal(map[string]string{"private_key": priv, "public_key": pub, "password_hash": string(utils.DSha256(password))})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	log.Debug("%v", json)
	return string(json), nil
}
