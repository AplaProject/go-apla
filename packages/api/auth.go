// MIT License
//
// Copyright (c) 2016 GenesisKernel
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
