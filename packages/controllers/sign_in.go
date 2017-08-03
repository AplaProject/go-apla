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

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const aSignIn = `ajax_sign_in`

// SignInJSON is a structure for the result of the sign in request
type SignInJSON struct {
	Address string `json:"address"`
	Result  bool   `json:"result"`
	Error   string `json:"error"`
}

func init() {
	newPage(aSignIn, `json`)
}

// AjaxSignIn checks signed uid
func (c *Controller) AjaxSignIn() interface{} {
	var result SignInJSON

	//	ret := `{"result":0}`
	c.r.ParseForm()
	key := c.r.FormValue("key")
	bkey, err := hex.DecodeString(key)
	stateID := converter.StrToInt64(c.r.FormValue("state_id"))
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if utils.PrivCountry && utils.OneCountry > 0 {
		stateID = utils.OneCountry
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

	if verify, _ := utils.CheckSign([][]byte{bkey}, msg, sign, true); !verify {
		result.Error = "incorrect signature"
		return result
	}
	result.Address = crypto.KeyToAddress(bkey)

	log.Debug("address : %s", result.Address)
	log.Debug("c.r.RemoteAddr %s", c.r.RemoteAddr)
	log.Debug("c.r.Header.Get(User-Agent) %s", c.r.Header.Get("User-Agent"))

	publicKey := []byte(key)
	walletID := int64(crypto.Address(publicKey))
	if err != nil {
		result.Error = err.Error()
		return result
	}
	citizen := &model.Citizen{}
	log.Debug("wallet_id : %d", walletID)
	if stateID > 0 {
		log.Debug("stateId %v", stateID)
		systemState := &model.SystemState{}
		err := systemState.Get(stateID)
		if err != nil {
			result.Error = err.Error()
			return result
		}
		citizen.SetTablePrefix(converter.Int64ToStr(stateID))
		err = citizen.Get(stateID)
		if err != nil {
			result.Error = err.Error()
			return result
		}
		log.Debug("citizenID %v", citizen.ID)
		if citizen.ID == 0 {
			stateID = 0
			if utils.PrivCountry {
				result.Error = "not a citizen"
				return result
			}
		}
	}
	result.Result = true
	c.sess.Set("wallet_id", walletID)
	c.sess.Set("address", result.Address)
	c.sess.Set("citizen_id", string(citizen.ID))
	c.sess.Set("state_id", stateID)
	log.Debug("wallet_id %d citizen_id %d state_id %d", walletID, citizen.ID, stateID)
	return result
}
