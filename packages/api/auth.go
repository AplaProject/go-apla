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
	"strings"

	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/crypto"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

var (
	jwtSecret = crypto.RandSeq(15)
)

// JWTClaims is storing jwt claims
type JWTClaims struct {
	UID         string `json:"uid,omitempty"`
	EcosystemID string `json:"ecosystem_id,omitempty"`
	KeyID       string `json:"key_id,omitempty"`
	jwt.StandardClaims
}

func jwtToken(r *http.Request) (*jwt.Token, error) {
	auth := r.Header.Get(`Authorization`)
	if len(auth) == 0 {
		return nil, nil
	}
	if strings.HasPrefix(auth, jwtPrefix) {
		auth = auth[len(jwtPrefix):]
	} else {
		return nil, fmt.Errorf(`wrong authorization value`)
	}
	return jwt.ParseWithClaims(auth, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
}

func jwtGenerateToken(w http.ResponseWriter, claims JWTClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func authWallet(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	if data.keyId == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("wallet is empty")
		return errorAPI(w, `E_UNAUTHORIZED`, http.StatusUnauthorized)
	}
	return nil
}

func authState(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	if data.keyId == 0 || data.ecosystemId <= 1 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("state is empty")
		return errorAPI(w, `E_UNAUTHORIZED`, http.StatusUnauthorized)
	}
	return nil
}
