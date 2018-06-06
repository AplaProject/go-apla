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
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

const tokenExpireMsg = "token is expired by"

var (
	jwtSecret       = []byte(crypto.RandSeq(15))
	jwtPrefix       = "Bearer "
	jwtExpire int64 = 36000 // By default, seconds

	authHeader = "AUTHORIZATION"

	errJWTAuthValue      = errors.New("wrong authorization value")
	errEcosystemNotFound = errors.New("ecosystem not found")
)

// JWTClaims is storing jwt claims
type JWTClaims struct {
	UID         string `json:"uid,omitempty"`
	EcosystemID string `json:"ecosystem_id,omitempty"`
	KeyID       string `json:"key_id,omitempty"`
	RoleID      string `json:"role_id,omitempty"`
	IsMobile    string `json:"is_mobile,omitempty"`
	jwt.StandardClaims
}

func TokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := getLogger(r)

		token, err := parseJWTToken(r.Header.Get(authHeader))
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.JWTError, "error": err}).Error("starting session")
			if err, ok := err.(jwt.ValidationError); ok {
				if (err.Errors & jwt.ValidationErrorExpired) != 0 {
					errorResponse(w, errTokenExpired, http.StatusUnauthorized, err.Error())
					return
				}
			}
			errorResponse(w, err, http.StatusBadRequest)
			return
		}

		if token != nil && token.Valid {
			r = setToken(r, token)
		}

		next.ServeHTTP(w, r)
	})
}

func ClientMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := getToken(r)

		var client *Client
		if token != nil { // get client from token
			var err error
			if client, err = getClientFromToken(token); err != nil {
				errorResponse(w, errServer, http.StatusInternalServerError)
				return
			}
		}
		if client == nil {
			// create client with default ecosystem
			client = &Client{EcosystemID: 1}
		}
		r = setClient(r, client)

		next.ServeHTTP(w, r)
	})
}

func AuthRequire(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		client := getClient(r)
		if client != nil && client.KeyID != 0 {
			next(w, r)
			return
		}

		logger := getLogger(r)
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("wallet is empty")
		errorResponse(w, errUnauthorized, http.StatusUnauthorized)
	}
}

func parseJWTToken(header string) (*jwt.Token, error) {
	if len(header) == 0 {
		return nil, nil
	}

	if strings.HasPrefix(header, jwtPrefix) {
		header = header[len(jwtPrefix):]
	} else {
		return nil, errJWTAuthValue
	}

	return jwt.ParseWithClaims(header, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
}

func getClientFromToken(token *jwt.Token) (*Client, error) {
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, nil
	}
	if len(claims.KeyID) == 0 {
		return nil, nil
	}

	ecosystemID := converter.StrToInt64(claims.EcosystemID)
	ecosystem := &model.Ecosystem{}
	found, err := ecosystem.Get(ecosystemID)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on getting ecosystem from db")
		return nil, err
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound, "id": ecosystemID, "error": errEcosystemNotFound}).Error("ecosystem not found")
		return nil, err
	}

	return &Client{
		EcosystemID:   converter.StrToInt64(claims.EcosystemID),
		EcosystemName: ecosystem.Name,
		KeyID:         converter.StrToInt64(claims.KeyID),
		IsMobile:      claims.IsMobile,
		RoleID:        converter.StrToInt64(claims.RoleID),
	}, nil
}

func generateJWTToken(claims JWTClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
