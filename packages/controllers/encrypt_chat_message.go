package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"

	"crypto/rand"
	"crypto/rsa"
	"errors"
)

func (c *Controller) EncryptChatMessage() (string, error) {

	var err error

	c.r.ParseForm()

	receiver := c.r.FormValue("receiver")
	message := c.r.FormValue("message")
	if len(message) > 5120 {
		return "", errors.New("incorrect message")
	}

	publicKey, err := c.Single("SELECT public_key_0 FROM users WHERE user_id  =  ?", receiver).Bytes()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	if len(publicKey) > 0 {
		pub, err := utils.BinToRsaPubKey(publicKey)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		enc_, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(message))
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		return utils.JsonAnswer(string(utils.BinToHex(enc_)), "success").String(), nil
	} else {
		return utils.JsonAnswer("Incorrect user_id", "error").String(), nil
	}

}
