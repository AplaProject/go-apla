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

package api_v2

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"
	"github.com/dgrijalva/jwt-go"
	hr "github.com/julienschmidt/httprouter"
	"github.com/op/go-logging"
	"gopkg.in/vmihailenco/msgpack.v2"
)

const (
	jwtPrefix = "Bearer "
	jwtExpire = 36000 // By default, seconds
)

type apiData struct {
	status int
	result interface{}
	params map[string]interface{}
	state  int64
	wallet int64
	token  *jwt.Token
	//	sess   session.SessionStore
}

type forSign struct {
	Time    string `json:"time"`
	ForSign string `json:"forsign"`
}

type hashTx struct {
	Hash string `json:"hash"`
}

const (
	pInt64 = iota
	pHex
	pString

	pOptional = 0x100
)

type apiHandle func(http.ResponseWriter, *http.Request, *apiData) error

var (
	log = logging.MustGetLogger("api")
)

func errorAPI(w http.ResponseWriter, msg string, code int) error {
	//	http.Error(w, fmt.Sprintf(`{"error": %q}`, msg), code)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintln(w, fmt.Sprintf(`{"error": %q}`, msg))
	return fmt.Errorf(msg)
}

func getPrefix(data *apiData) (prefix string) {
	return converter.Int64ToStr(data.state)
}

func getSignHeader(txName string, data *apiData) tx.Header {
	return tx.Header{Type: int(utils.TypeInt(txName)), Time: time.Now().Unix(),
		UserID: data.state, StateID: data.wallet}
}

func getHeader(txName string, data *apiData) (tx.Header, error) {
	publicKey := []byte("null")
	if _, ok := data.params[`pubkey`]; ok && len(data.params[`pubkey`].([]byte)) > 0 {
		publicKey = data.params[`pubkey`].([]byte)
		lenpub := len(publicKey)
		if lenpub > 64 {
			publicKey = publicKey[lenpub-64:]
		}
	}
	signature := data.params[`signature`].([]byte)
	if len(signature) == 0 {
		return tx.Header{}, fmt.Errorf("signature is empty")
	}
	return tx.Header{Type: int(utils.TypeInt(txName)), Time: converter.StrToInt64(data.params[`time`].(string)),
		UserID: data.wallet, StateID: data.state, PublicKey: publicKey,
		BinSignatures: converter.EncodeLengthPlusData(signature)}, nil
}

func sendEmbeddedTx(txType int, userID int64, toSerialize interface{}) (*hashTx, error) {
	var hash []byte
	serializedData, err := msgpack.Marshal(toSerialize)
	if err != nil {
		return nil, err
	}
	if hash, err = model.SendTx(int64(txType), userID,
		append(converter.DecToBin(int64(txType), 1), serializedData...)); err != nil {
		return nil, err
	}
	return &hashTx{Hash: string(hash)}, nil
}

// DefaultHandler is a common handle function for api requests
func DefaultHandler(params map[string]int, handlers ...apiHandle) hr.Handle {
	return hr.Handle(func(w http.ResponseWriter, r *http.Request, ps hr.Params) {
		var (
			err  error
			data apiData
		)
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("API Recovered", fmt.Sprintf("%s: %s", r, debug.Stack()))
				errorAPI(w, "", http.StatusInternalServerError)
				log.Error("API Recovered", r)
			}
		}()
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		token, err := jwtToken(r)
		if err != nil {
			errorAPI(w, err.Error(), http.StatusBadRequest)
			return
		}
		data.token = token
		if token != nil && token.Valid {
			if claims, ok := token.Claims.(*JWTClaims); ok && len(claims.Wallet) > 0 {
				data.state = converter.StrToInt64(claims.State)
				data.wallet = converter.StrToInt64(claims.Wallet)
				w.Header().Set("Authorization", jwtPrefix+token.Raw)
			}
		}
		// Getting and validating request parameters
		r.ParseForm()
		data.params = make(map[string]interface{})
		for _, par := range ps {
			data.params[par.Key] = par.Value
		}
		for key, par := range params {
			val := r.FormValue(key)
			if par&pOptional == 0 && len(val) == 0 {
				errorAPI(w, fmt.Sprintf(`Value %s is undefined`, key), http.StatusBadRequest)
				return
			}
			switch par & 0xff {
			case pInt64:
				data.params[key] = converter.StrToInt64(val)
			case pHex:
				bin, err := hex.DecodeString(val)
				if err != nil {
					errorAPI(w, err.Error(), http.StatusBadRequest)
					return
				}
				data.params[key] = bin
			case pString:
				data.params[key] = val
			}
		}
		for _, handler := range handlers {
			if handler(w, r, &data) != nil {
				return
			}
		}
		jsonResult, err := json.Marshal(data.result)
		if err != nil {
			errorAPI(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(jsonResult)
	})
}
