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
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

type refreshResult struct {
	Token   string `json:"token,omitempty"`
	Refresh string `json:"refresh,omitempty"`
}

func refresh(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {

	if data.token == nil || !data.token.Valid {
		if data.token == nil {
			logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("token is nil")
		}
		if !data.token.Valid {
			logger.WithFields(log.Fields{"type": consts.InvalidObject}).Error("token is invalid")
		}
		return errorAPI(w, `E_TOKEN`, http.StatusBadRequest)
	}
	claims, ok := data.token.Claims.(*JWTClaims)
	if !ok || converter.StrToInt64(claims.KeyID) == 0 {
		logger.WithFields(log.Fields{"type": consts.JWTError}).Error("getting jwt claims")
		return errorAPI(w, `E_TOKEN`, http.StatusBadRequest)
	}
	token, err := jwt.ParseWithClaims(data.params[`token`].(string), &JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				logger.WithFields(log.Fields{"type": consts.JWTError, "signing_method": token.Header["alg"]}).Error("unexpected signing method")
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.JWTError, "signing_method": token.Header["alg"]}).Error("unexpected signing method")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	if token == nil || !token.Valid {
		if data.token == nil {
			logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("token is nil")
		}
		if !token.Valid {
			logger.WithFields(log.Fields{"type": consts.InvalidObject}).Error("token is invalid")
		}
		return errorAPI(w, `E_REFRESHTOKEN`, http.StatusBadRequest)
	}
	refClaims, ok := token.Claims.(*JWTClaims)
	if !ok || refClaims.KeyID != claims.KeyID || refClaims.EcosystemID != claims.EcosystemID {
		logger.WithFields(log.Fields{"type": consts.JWTError}).Error("token wallet or state is invalid")
		return errorAPI(w, `E_REFRESHTOKEN`, http.StatusBadRequest)
	}
	var result refreshResult
	data.result = &result

	expire := data.params[`expire`].(int64)
	if expire == 0 {
		logger.Warning("expire is 0, using jwt expire")
		expire = jwtExpire
	}
	claims.StandardClaims.ExpiresAt = time.Now().Add(time.Second * time.Duration(expire)).Unix()
	result.Token, err = jwtGenerateToken(w, *claims)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.JWTError, "error": err}).Error("generating jwt token")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	claims.StandardClaims.ExpiresAt = time.Now().Add(time.Hour * 30 * 24).Unix()
	result.Refresh, err = jwtGenerateToken(w, *claims)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.JWTError, "error": err}).Error("generating jwt token")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	return nil
}
