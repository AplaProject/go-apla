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

package apiv2

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AplaProject/go-apla/packages/converter"

	"github.com/dgrijalva/jwt-go"
)

type refreshResult struct {
	Token   string `json:"token,omitempty"`
	Refresh string `json:"refresh,omitempty"`
}

func refresh(w http.ResponseWriter, r *http.Request, data *apiData) error {

	if data.token == nil || !data.token.Valid {
		return errorAPI(w, `E_TOKEN`, http.StatusBadRequest)
	}
	claims, ok := data.token.Claims.(*JWTClaims)
	if !ok || converter.StrToInt64(claims.KeyID) == 0 {
		return errorAPI(w, `E_TOKEN`, http.StatusBadRequest)
	}
	token, err := jwt.ParseWithClaims(data.params[`token`].(string), &JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})
	if err != nil {
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	if token == nil || !token.Valid {
		return errorAPI(w, `E_REFRESHTOKEN`, http.StatusBadRequest)
	}
	refClaims, ok := token.Claims.(*JWTClaims)
	if !ok || refClaims.KeyID != claims.KeyID || refClaims.EcosystemID != claims.EcosystemID {
		return errorAPI(w, `E_REFRESHTOKEN`, http.StatusBadRequest)
	}
	var result refreshResult
	data.result = &result

	expire := data.params[`expire`].(int64)
	if expire == 0 {
		expire = jwtExpire
	}
	claims.StandardClaims.ExpiresAt = time.Now().Add(time.Second * time.Duration(expire)).Unix()
	result.Token, err = jwtGenerateToken(w, *claims)
	if err != nil {
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	claims.StandardClaims.ExpiresAt = time.Now().Add(time.Hour * 30 * 24).Unix()
	result.Refresh, err = jwtGenerateToken(w, *claims)
	if err != nil {
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	return nil
}
