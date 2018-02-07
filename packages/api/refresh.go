// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"

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
		} else if !data.token.Valid {
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
