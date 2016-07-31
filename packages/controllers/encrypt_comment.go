package controllers

import (
	"encoding/json"
	"github.com/DayLightProject/go-daylight/packages/utils"

	"crypto/rand"
	"crypto/rsa"
	"errors"
)

func (c *Controller) EncryptComment() (string, error) {

	var err error

	c.r.ParseForm()

	txType := c.r.FormValue("type")
	var toId int64
	var toIds []int64
	toIds_ := c.r.FormValue("to_ids")
	if len(toIds_) == 0 {
		toId = utils.StrToInt64(c.r.FormValue("to_id"))
	} else {
		var toIdsMap map[string]string
		err = json.Unmarshal([]byte(toIds_), &toIdsMap)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		for _, uid := range toIdsMap {
			if utils.StrToInt64(uid) > 0 {
				toIds = append(toIds, utils.StrToInt64(uid))
			}
		}
	}

	comment := c.r.FormValue("comment")
	if len(comment) > 1024 {
		return "", errors.New("incorrect comment")
	}

	var toUserId int64
	if txType == "project" {
		toUserId, err = c.Single("SELECT user_id FROM cf_projects WHERE id  =  ?", toId).Int64()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	} else {
		toUserId = toId
	}

	if len(toIds) == 0 {
		toIds = []int64{toUserId}
	}

	log.Debug("toId:", toId)
	log.Debug("toIds:", toIds)
	log.Debug("toUserId:", toUserId)
	enc := make(map[string]string)
	for i := 0; i < len(toIds); i++ {
		if toIds[i] == 0 {
			enc[utils.IntToStr(i)] = "0"
			continue
		}
		// если получатель майнер, тогда шифруем нодовским ключем
		minersData, err := c.OneRow("SELECT miner_id, node_public_key FROM miners_data WHERE user_id  =  ?", toIds[i]).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		var publicKey string
		if utils.StrToInt(minersData["miner_id"]) > 0 && txType != "cash_request" && txType != "bug_reporting" && txType != "project" && txType != "money_back" && txType != "restoringAccess" {
			publicKey = minersData["node_public_key"]
		} else {
			publicKey, err = c.Single("SELECT public_key_0 FROM users WHERE user_id  =  ?", toIds[i]).String()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		}

		pub, err := utils.BinToRsaPubKey([]byte(publicKey))
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		enc_, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(comment))
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		enc[utils.IntToStr(i)] = string(utils.BinToHex(enc_))
	}
	if len(enc) < 5 && len(enc) > 0 {
		for i := len(enc); i < 5; i++ {
			enc[utils.IntToStr(i)] = "0"
		}
	}
	log.Debug("enc:", enc)
	if txType != "arbitration_arbitrators" {
		return string(enc["0"]), nil
	} else {
		result, err := json.Marshal(enc)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		return string(result), nil
	}
}
