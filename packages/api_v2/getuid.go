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
	"math/rand"
	"net/http"
	"time"

	//	"github.com/EGaaS/go-egaas-mvp/packages/config"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	//	"github.com/EGaaS/go-egaas-mvp/packages/model"

	"github.com/dgrijalva/jwt-go"
)

var (
	installed bool
)

type getUIDResult struct {
	UID     string `json:"uid,omitempty"`
	Token   string `json:"token,omitempty"`
	Expire  string `json:"expire,omitempty"`
	State   string `json:"state,omitempty"`
	Wallet  string `json:"wallet,omitempty"`
	Address string `json:"address,omitempty"`
}

// If State == 0 then APLA has not been installed
// If Wallet == 0 then login is required

func getUID(w http.ResponseWriter, r *http.Request, data *apiData) (err error) {
	var result getUIDResult

	data.result = &result

	if data.token != nil && data.token.Valid {
		if claims, ok := data.token.Claims.(*JWTClaims); ok && len(claims.Wallet) > 0 {
			result.State = claims.State
			result.Expire = converter.Int64ToStr(claims.ExpiresAt - time.Now().Unix())
			result.Wallet = claims.Wallet
			return nil
		}
	}
	/*	if !installed {
		if model.DBConn == nil && !config.IsExist() {
			return nil
		}
		installed = true
	}*/
	result.UID = converter.Int64ToStr(rand.New(rand.NewSource(time.Now().Unix())).Int63())
	claims := JWTClaims{
		UID: result.UID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * 5).Unix(),
		},
	}
	result.Token, err = jwtGenerateToken(w, claims)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	return
}
