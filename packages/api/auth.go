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
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/types"

	"github.com/dgrijalva/jwt-go"
)

var (
	jwtSecret = []byte(crypto.RandSeq(15))
	jwtPrefix = "Bearer "
	jwtExpire = 36000 // By default, seconds

	errJWTAuthValue      = errors.New("wrong authorization value")
	errEcosystemNotFound = errors.New("ecosystem not found")
)

// JWTClaims is storing jwt claims
type JWTClaims struct {
	UID         string `json:"uid,omitempty"`
	EcosystemID string `json:"ecosystem_id,omitempty"`
	KeyID       string `json:"key_id,omitempty"`
	AccountID   string `json:"account_id,omitempty"`
	RoleID      string `json:"role_id,omitempty"`
	IsMobile    bool   `json:"is_mobile,omitempty"`
	jwt.StandardClaims
}

func generateJWTToken(claims JWTClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
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

func getClientFromToken(token *jwt.Token, ecosysNameService types.EcosystemNameGetter) (*Client, error) {
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, nil
	}
	if len(claims.KeyID) == 0 {
		return nil, nil
	}

	client := &Client{
		EcosystemID: converter.StrToInt64(claims.EcosystemID),
		KeyID:       converter.StrToInt64(claims.KeyID),
		AccountID:   claims.AccountID,
		IsMobile:    claims.IsMobile,
		RoleID:      converter.StrToInt64(claims.RoleID),
	}

	sID := converter.StrToInt64(claims.EcosystemID)
	name, err := ecosysNameService.GetEcosystemName(sID)
	if err != nil {
		return nil, err
	}

	client.EcosystemName = name
	return client, nil
}

type authStatusResponse struct {
	IsActive  bool  `json:"active"`
	ExpiresAt int64 `json:"exp,omitempty"`
}

func getAuthStatus(w http.ResponseWriter, r *http.Request) {
	result := new(authStatusResponse)
	defer jsonResponse(w, result)

	token := getToken(r)
	if token == nil {
		return
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return
	}

	result.IsActive = true
	result.ExpiresAt = claims.ExpiresAt
}
