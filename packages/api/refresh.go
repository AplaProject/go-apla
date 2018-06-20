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
	"net/http"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"

	log "github.com/sirupsen/logrus"
)

type refreshResult struct {
	Token   string `json:"token,omitempty"`
	Refresh string `json:"refresh,omitempty"`
}

type refreshForm struct {
	form
	Token  string `schema:"token"`
	Expire int64  `schema:"expire"`
}

func refreshHandler(w http.ResponseWriter, r *http.Request) {
	form := &refreshForm{}
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err)
		return
	}

	claims, err := getRefreshTokenClaims(r, form.Token)
	if err != nil {
		errorResponse(w, err)
		return
	}

	if _, err := getAccount(r, converter.StrToInt64(claims.EcosystemID), converter.StrToInt64(claims.KeyID)); err != nil {
		errorResponse(w, err)
		return
	}

	result := &refreshResult{}

	logger := getLogger(r)
	if form.Expire == 0 {
		logger.Warning("expire is 0, using jwt expire")
		form.Expire = jwtExpire
	}
	claims.StandardClaims.ExpiresAt = time.Now().Add(time.Second * time.Duration(form.Expire)).Unix()

	result.Token, err = generateJWTToken(*claims)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.JWTError, "error": err}).Error("generating jwt token")
		errorResponse(w, err)
		return
	}
	claims.StandardClaims.ExpiresAt = time.Now().Add(time.Hour * 30 * 24).Unix()
	result.Refresh, err = generateJWTToken(*claims)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.JWTError, "error": err}).Error("generating jwt token")
		errorResponse(w, err)
		return
	}

	jsonResponse(w, result)
}

func getRefreshTokenClaims(r *http.Request, val string) (*JWTClaims, error) {
	logger := getLogger(r)

	token := getToken(r)
	if token == nil {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("token is nil")
		return nil, errToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || len(claims.KeyID) == 0 {
		logger.WithFields(log.Fields{"type": consts.JWTError}).Error("getting jwt claims")
		return nil, errToken
	}

	refToken, err := parseJWTToken(val)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.JWTError, "signing_method": token.Header["alg"]}).Error("unexpected signing method")
		return nil, err
	}

	if refToken == nil || !refToken.Valid {
		logger.WithFields(log.Fields{"type": consts.InvalidObject}).Error("token is invalid")
		return nil, errRefreshToken
	}
	refClaims, ok := refToken.Claims.(*JWTClaims)
	if !ok || refClaims.KeyID != claims.KeyID || refClaims.EcosystemID != claims.EcosystemID {
		logger.WithFields(log.Fields{"type": consts.JWTError}).Error("token wallet or state is invalid")
		return nil, errRefreshToken
	}

	return refClaims, nil
}
