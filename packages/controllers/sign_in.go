// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package controllers

import (
	"encoding/hex"

	"github.com/DayLightProject/go-daylight/packages/lib"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

const ASignIn = `ajax_sign_in`

type SignInJson struct {
	Address string `json:"address"`
	Result  bool   `json:"result"`
	Error   string `json:"error"`
}

func init() {
	newPage(ASignIn, `json`)
}

func (c *Controller) AjaxSignIn() interface{} {
	var result SignInJson

	//	ret := `{"result":0}`
	c.r.ParseForm()
	key := c.r.FormValue("key")
	bkey, err := hex.DecodeString(key)
	stateId := utils.StrToInt64(c.r.FormValue("state_id"))
	if err != nil {
		result.Error = err.Error()
		return result
	}
	sign, _ := hex.DecodeString(c.r.FormValue("sign"))
	var msg string
	switch uid := c.sess.Get(`uid`).(type) {
	case string:
		msg = uid
	default:
		result.Error = "unknown uid"
		return result
	}

	if verify, _ := utils.CheckECDSA([][]byte{bkey}, msg, sign, true); !verify {
		result.Error = "incorrect signature"
		return result
	}
	result.Result = true
	result.Address = lib.KeyToAddress(bkey)
	c.sess.Set("address", result.Address)
	log.Debug("c.r.RemoteAddr %s", c.r.RemoteAddr)
	log.Debug("c.r.Header.Get(User-Agent) %s", c.r.Header.Get("User-Agent"))

	publicKey := []byte(key)
	walletId, err := c.GetWalletIdByPublicKey(publicKey)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	/*	err = c.ExecSql("UPDATE config SET dlt_wallet_id = ?", walletId)
		if err != nil {
			result.Error = err.Error()
			return result
		}*/
	c.sess.Set("wallet_id", walletId)
	var citizenId int64
	if stateId > 0 {
		result = SignInJson{}
		log.Debug("stateId %v", stateId)
		if _, err := c.CheckStateName(stateId); err == nil {
			citizenId, err = c.Single(`SELECT id FROM "`+utils.Int64ToStr(stateId)+`_citizens" WHERE hex(public_key_0) = ?`,
				string(publicKey)).Int64()
			if err != nil {
				result.Error = err.Error()
				return result
			}
			log.Debug("citizenId %v", citizenId)
			if citizenId == 0 {
				stateId = 0
				result.Error = "not a citizen"
			} else {
				result.Result = true
				result.Address = lib.KeyToAddress(bkey)
			}
		} else {
			result.Error = err.Error()
			return result
		}
	}
	/*	citizenId, err := c.GetCitizenIdByPublicKey(publicKey)
		err = c.ExecSql("UPDATE config SET citizen_id = ?", citizenId)
		if err != nil {
			result.Error = err.Error()
			return result
		}*/
	c.sess.Set("citizen_id", citizenId)
	c.sess.Set("state_id", stateId)
	return result //`{"result":1,"address": "` + address + `"}`, nil
}
