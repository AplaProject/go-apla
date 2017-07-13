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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/sql"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"
	"github.com/astaxie/beego/session"
	hr "github.com/julienschmidt/httprouter"
	"github.com/op/go-logging"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type apiData struct {
	status int
	result interface{}
	params map[string]interface{}
	sess   session.SessionStore
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
	log     = logging.MustGetLogger("api")
	apiSess *session.Manager
)

// SetSession must be called for assigning session
func SetSession(s *session.Manager) {
	apiSess = s
}

func errorAPI(w http.ResponseWriter, msg string, code int) error {
	http.Error(w, msg, code)
	return fmt.Errorf(msg)
}

func getPrefix(data *apiData) (prefix string) {
	if glob, ok := data.params[`global`].(int64); ok && glob > 0 {
		prefix = `global`
	} else {
		prefix = converter.Int64ToStr(data.sess.Get(`state`).(int64))
	}
	return
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
	var stateID int64
	userID := data.sess.Get(`wallet`).(int64)
	if data.sess.Get(`state`) != nil {
		stateID = data.sess.Get(`state`).(int64)
	}
	return tx.Header{Type: int(utils.TypeInt(txName)), Time: converter.StrToInt64(data.params[`time`].(string)),
		UserID: userID, StateID: stateID, PublicKey: publicKey,
		BinSignatures: converter.EncodeLengthPlusData(signature)}, nil
}

func getForSign(txName string, data *apiData, append string) *forSign {
	timeNow := time.Now().Unix()
	forsign := fmt.Sprintf(`%d,%d,%d,%d,`, utils.TypeInt(txName), timeNow, data.sess.Get(`citizen`).(int64),
		data.sess.Get(`state`).(int64))
	return &forSign{Time: converter.Int64ToStr(timeNow), ForSign: forsign + append}
}

func sendEmbeddedTx(txType int, userID int64, toSerialize interface{}) (*hashTx, error) {
	var hash []byte
	serializedData, err := msgpack.Marshal(toSerialize)
	if err != nil {
		return nil, err
	}
	if hash, err = sql.DB.SendTx(int64(txType), userID,
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
				log.Error("API Recovered", r)
			}
		}()
		if apiSess == nil {
			errorAPI(w, `Session is undefined`, http.StatusConflict)
			return
		}

		data.sess, err = apiSess.SessionStart(w, r)
		if err != nil {
			errorAPI(w, err.Error(), http.StatusConflict)
			return
		}
		defer data.sess.SessionRelease(w)

		// Getting and validating request parameters
		r.ParseForm()
		data.params = make(map[string]interface{})
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
		for _, par := range ps {
			data.params[par.Key] = par.Value
		}
		for _, handler := range handlers {
			if handler(w, r, &data) != nil {
				return
			}
		}
		jsonResult, err := json.Marshal(data.result)
		if err != nil {
			errorAPI(w, err.Error(), http.StatusConflict)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(jsonResult)
	})
}
