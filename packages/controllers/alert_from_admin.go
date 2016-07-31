package controllers

import (
	"crypto"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/mcuadros/go-version"
)

type alertType struct {
	Message   map[string]string
	Signature string
}

func (c *Controller) AlertFromAdmin() (string, error) {

	if c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	alertMessage := ""
	alert, err := utils.GetHttpTextAnswer("http://dcoin.club/alert.json")
	if len(alert) > 0 {
		alertData := new(alertType)
		err = json.Unmarshal([]byte(alert), &alertData)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
		}

		messageJson, err := json.Marshal(alertData.Message)
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
		}

		pub, err := utils.BinToRsaPubKey(utils.HexToBin(consts.ALERT_KEY))
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
		}
		err = rsa.VerifyPKCS1v15(pub, crypto.SHA1, utils.HashSha1(string(messageJson)), []byte(utils.HexToBin(alertData.Signature)))
		if err != nil {
			log.Error("%v", utils.ErrInfo(err))
		}

		if version.Compare(alertData.Message["version"], consts.VERSION, ">") {
			alertMessage = alertData.Message[utils.Int64ToStr(c.LangInt)]
			return utils.JsonAnswer(alertMessage, "success").String(), nil
		}
	}
	return ``, nil
}
