// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package api

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

const jwtUIDExpire = time.Second * 5

type getUIDResult struct {
	UID         string `json:"uid,omitempty"`
	Token       string `json:"token,omitempty"`
	Expire      string `json:"expire,omitempty"`
	EcosystemID string `json:"ecosystem_id,omitempty"`
	KeyID       string `json:"key_id,omitempty"`
	Address     string `json:"address,omitempty"`
	NetworkID   string `json:"network_id,omitempty"`
}

func getUIDHandler(w http.ResponseWriter, r *http.Request) {
	result := new(getUIDResult)
	result.NetworkID = converter.Int64ToStr(conf.Config.NetworkID)
	token := getToken(r)
	if token != nil {
		if claims, ok := token.Claims.(*JWTClaims); ok && len(claims.KeyID) > 0 {
			result.EcosystemID = claims.EcosystemID
			result.Expire = converter.Int64ToStr(claims.ExpiresAt - time.Now().Unix())
			result.KeyID = claims.KeyID
			jsonResponse(w, result)
			return
		}
	}

	result.UID = converter.Int64ToStr(rand.New(rand.NewSource(time.Now().Unix())).Int63())
	claims := JWTClaims{
		UID:         result.UID,
		EcosystemID: "1",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(jwtUIDExpire).Unix(),
		},
	}

	var err error
	if result.Token, err = generateJWTToken(claims); err != nil {
		logger := getLogger(r)
		logger.WithFields(log.Fields{"type": consts.JWTError, "error": err}).Error("generating jwt token")
		errorResponse(w, err)
		return
	}

	jsonResponse(w, result)
}
