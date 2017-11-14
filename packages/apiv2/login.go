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
	"net/http"
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/notificator"
	"github.com/AplaProject/go-apla/packages/publisher"

	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

type loginResult struct {
	Token       string `json:"token,omitempty"`
	Refresh     string `json:"refresh,omitempty"`
	EcosystemID string `json:"ecosystem_id,omitempty"`
	KeyID       string `json:"key_id,omitempty"`
	Address     string `json:"address,omitempty"`
	NotifyKey   string `json:"notify_key,omitempty"`
}

func login(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var (
		pubkey []byte
		wallet int64
		msg    string
		err    error
	)

	if data.token != nil && data.token.Valid {
		if claims, ok := data.token.Claims.(*JWTClaims); ok {
			msg = claims.UID
		}
	}
	if len(msg) == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("UID is empty")
		return errorAPI(w, `E_UNKNOWNUID`, http.StatusBadRequest)
	}
	state := data.ecosystemId
	if data.params[`ecosystem`].(int64) > 0 {
		state = data.params[`ecosystem`].(int64)
	}
	if state == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Warning("state is empty, using 1 as a state")
		state = 1
	}
	if len(data.params[`key_id`].(string)) > 0 {
		wallet = converter.StringToAddress(data.params[`key_id`].(string))
	} else if len(data.params[`pubkey`].([]byte)) > 0 {
		wallet = crypto.Address(data.params[`pubkey`].([]byte))
	}
	pubkey, err = model.Single(`select pub from "`+converter.Int64ToStr(state)+`_keys" where id=?`, wallet).Bytes()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting public key from keys")
		return errorAPI(w, err, http.StatusBadRequest)
	}
	if state > 1 && len(pubkey) == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("public key is empty, and state is not default")
		return errorAPI(w, `E_STATELOGIN`, http.StatusForbidden, wallet, state)
	}
	if len(pubkey) == 0 {
		pubkey = data.params[`pubkey`].([]byte)
		if len(pubkey) == 0 {
			logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("public key is empty")
			return errorAPI(w, `E_EMPTYPUBLIC`, http.StatusBadRequest)
		}
	}
	verify, err := crypto.CheckSign(pubkey, msg, data.params[`signature`].([]byte))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.CryptoError, "pubkey": pubkey, "msg": msg, "signature": string(data.params["signature"].([]byte))}).Error("checking signature")
		return errorAPI(w, err, http.StatusBadRequest)
	}
	if !verify {
		logger.WithFields(log.Fields{"type": consts.InvalidObject, "pubkey": pubkey, "msg": msg, "signature": string(data.params["signature"].([]byte))}).Error("incorrect signature")
		return errorAPI(w, `E_SIGNATURE`, http.StatusBadRequest)
	}
	address := crypto.KeyToAddress(pubkey)
	result := loginResult{EcosystemID: converter.Int64ToStr(state), KeyID: converter.Int64ToStr(wallet),
		Address: address}
	data.result = &result
	expire := data.params[`expire`].(int64)
	if expire == 0 {
		logger.WithFields(log.Fields{"type": consts.JWTError, "expire": jwtExpire}).Warning("using expire from jwt")
		expire = jwtExpire
	}
	claims := JWTClaims{
		KeyID:       result.KeyID,
		EcosystemID: result.EcosystemID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * time.Duration(expire)).Unix(),
		},
	}
	result.Token, err = jwtGenerateToken(w, claims)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.JWTError, "error": err}).Error("generating jwt token")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	claims.StandardClaims.ExpiresAt = time.Now().Add(time.Hour * 30 * 24).Unix()
	result.Refresh, err = jwtGenerateToken(w, claims)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.JWTError, "error": err}).Error("generating jwt token")
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	result.NotifyKey, err = publisher.GetHMACSign(wallet)
	if err != nil {
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	notificator.AddUser(wallet, state)

	return nil
}
