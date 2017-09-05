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

package api_v2

import (
	"fmt"
	"net/http"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/dgrijalva/jwt-go"
)

type loginResult struct {
	Token   string `json:"token,omitempty"`
	Refresh string `json:"refresh,omitempty"`
	State   string `json:"state,omitempty"`
	Wallet  string `json:"wallet,omitempty"`
	Address string `json:"address,omitempty"`
}

func login(w http.ResponseWriter, r *http.Request, data *apiData) error {
	var msg string

	if data.token != nil && data.token.Valid {
		if claims, ok := data.token.Claims.(*JWTClaims); ok {
			msg = claims.UID
		}
	}
	if len(msg) == 0 {
		return errorAPI(w, "unknown uid", http.StatusBadRequest)
	}
	pubkey := data.params[`pubkey`].([]byte)
	verify, err := crypto.CheckSign(pubkey, msg, data.params[`signature`].([]byte))
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusBadRequest)
	}
	if !verify {
		return errorAPI(w, `signature is incorrect`, http.StatusBadRequest)
	}
	state := data.params[`state`].(int64)
	if state == 0 {
		state = 1
	}
	address := crypto.KeyToAddress(pubkey)
	wallet := crypto.Address(pubkey)
	if state > 1 {
		sysState := &model.SystemState{}
		if exist, err := sysState.IsExists(state); err == nil && exist {
			citizen, err := model.Single(`SELECT id FROM "`+converter.Int64ToStr(state)+`_keys" WHERE id = ?`,
				wallet).Int64()
			if err != nil {
				return errorAPI(w, err.Error(), http.StatusInternalServerError)
			}
			if citizen == 0 {
				return errorAPI(w, fmt.Sprintf("not a membership of ecosystem %d", state), http.StatusForbidden)
			}
		} else {
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}
	}

	result := loginResult{State: converter.Int64ToStr(state), Wallet: converter.Int64ToStr(wallet),
		Address: address}
	data.result = &result
	claims := JWTClaims{
		Wallet: result.Wallet,
		State:  result.State,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * jwtExpire).Unix(),
		},
	}
	result.Token, err = jwtGenerateToken(w, claims)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	claims.StandardClaims.ExpiresAt = time.Now().Add(time.Hour * 30 * 24).Unix()
	result.Refresh, err = jwtGenerateToken(w, claims)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}

	return nil
}
