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

package api

import (
	"fmt"
	"net/http"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

type loginResult struct {
	Address string `json:"address"`
}

func login(w http.ResponseWriter, r *http.Request, data *apiData) error {
	logger.LogDebug(consts.FuncStarted, "")
	var msg string
	switch uid := data.sess.Get(`uid`).(type) {
	case string:
		msg = uid
	default:
		logger.LogError(consts.RouteError, data.sess.Get(`uid`))
		return errorAPI(w, "unknown uid", http.StatusBadRequest)
	}
	pubkey := data.params[`pubkey`].([]byte)
	verify, err := crypto.CheckSign(pubkey, msg, data.params[`signature`].([]byte))
	if err != nil {
		logger.LogError(consts.RouteError, err)
		return errorAPI(w, err.Error(), http.StatusBadRequest)
	}
	if !verify {
		logger.LogDebug(consts.RouteError, fmt.Sprintf("pubkey: %s, msg: %s, sign: %s", string(pubkey), msg, string(data.params[`signature`].([]byte))))
		return errorAPI(w, `signature is incorrect`, http.StatusBadRequest)
	}
	state := data.params[`state`].(int64)
	address := crypto.KeyToAddress(pubkey)
	wallet := crypto.Address(pubkey)
	var citizen int64
	if state > 0 {
		sysState := &model.SystemState{}
		if exist, err := sysState.IsExists(state); err == nil && exist {
			citizen, err = model.Single(`SELECT id FROM "`+converter.Int64ToStr(state)+`_citizens" WHERE id = ?`,
				wallet).Int64()
			if err != nil {
				logger.LogError(consts.DBError, err)
				return errorAPI(w, err.Error(), http.StatusInternalServerError)
			}
			if citizen == 0 {
				logger.LogInfo(consts.RecordNotFoundError, fmt.Sprintf("stateID: %d, citizenID: %d", state, wallet))
				state = 0
				if utils.PrivCountry {
					logger.LogError(consts.RouteError, "not a citizen")
					return errorAPI(w, "not a citizen", http.StatusForbidden)
				}
			}
		} else {
			logger.LogError(consts.RecordNotFoundError, fmt.Sprintf("state %d is not exists", wallet))
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}
	}

	data.result = &loginResult{Address: address}
	data.sess.Set("wallet", wallet)
	data.sess.Set("address", address)
	data.sess.Set("citizen", citizen)
	data.sess.Set("state", state)
	return nil
}
