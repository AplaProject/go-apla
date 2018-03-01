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

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/notificator"
	"github.com/GenesisKernel/go-genesis/packages/publisher"

	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"

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
	IsNode      bool   `json:"isnode,omitempty"`
	IsOwner     bool   `json:"isowner,omitempty"`
	IsVDE       bool   `json:"vde,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
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
	var (
		sp      model.StateParameter
		founder int64
	)
	sp.SetTablePrefix(converter.Int64ToStr(state))
	if ok, err := sp.Get(nil, "founder_account"); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting founder_account parameter")
		return errorAPI(w, `E_SERVER`, http.StatusBadRequest)
	} else if ok {
		founder = converter.StrToInt64(sp.Value)
	}

	result := loginResult{EcosystemID: converter.Int64ToStr(state), KeyID: converter.Int64ToStr(wallet),
		Address: address, IsOwner: founder == wallet, IsNode: conf.Config.KeyID == wallet,
		IsVDE: model.IsTable(fmt.Sprintf(`%d_vde_tables`, state))}
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
	result.NotifyKey, result.Timestamp, err = publisher.GetHMACSign(wallet)
	if err != nil {
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	notificator.AddUser(wallet, state)
	notificator.UpdateNotifications(state, []int64{wallet})

	return nil
}
