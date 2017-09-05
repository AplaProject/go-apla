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
	"net/http"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
)

type refreshResult struct {
	Token   string `json:"token,omitempty"`
	Refresh string `json:"refresh,omitempty"`
}

func refresh(w http.ResponseWriter, r *http.Request, data *apiData) error {

	if data.token == nil || !data.token.Valid {
		return errorAPI(w, `invalid token`, http.StatusBadRequest)
	}
	claims, ok := data.token.Claims.(*JWTClaims)
	if !ok || converter.StrToInt64(claims.Wallet) == 0 {
		return errorAPI(w, `invalid token`, http.StatusBadRequest)
	}
	/*
		refresh := data.params[`refresh`].(string)

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
	*/
	return nil
}
