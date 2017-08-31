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
	"strconv"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"
	"github.com/astaxie/beego/session"
	hr "github.com/julienschmidt/httprouter"
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

func getSignHeader(txName string, data *apiData) tx.Header {
	var stateID int64

	userID := data.sess.Get(`wallet`).(int64)
	if data.sess.Get(`state`) != nil {
		stateID = data.sess.Get(`state`).(int64)
	}
	return tx.Header{Type: int(utils.TypeInt(txName)), Time: time.Now().Unix(),
		UserID: userID, StateID: stateID}
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
	time, err := strconv.ParseInt(data.params["time"].(string), 10, 64)
	if err != nil {
		logger.LogInfo(consts.StrToIntError, data.params["time"])
	}
	return tx.Header{Type: int(utils.TypeInt(txName)), Time: time,
		UserID: userID, StateID: stateID, PublicKey: publicKey,
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
				errorAPI(w, "", http.StatusInternalServerError)
				logger.LogError(consts.PanicRecoveredError, r)
			}
		}()
		if apiSess == nil {
			errorAPI(w, `Session is undefined`, http.StatusForbidden)
			logger.LogDebug(consts.SessionError, "")
			return
		}

		data.sess, err = apiSess.SessionStart(w, r)
		if err != nil {
			errorAPI(w, err.Error(), http.StatusInternalServerError)
			logger.LogError(consts.SessionError, "")
			return
		}
		defer data.sess.SessionRelease(w)
		// Getting and validating request parameters
		r.ParseForm()
		data.params = make(map[string]interface{})
		for _, par := range ps {
			data.params[par.Key] = par.Value
		}
		for key, par := range params {
			val := r.FormValue(key)
			if par&pOptional == 0 && len(val) == 0 {
				errorMessage := fmt.Sprintf(`Value %s is undefined`, key)
				errorAPI(w, errorMessage, http.StatusBadRequest)
				logger.LogError(consts.RouteError, errorMessage)
				return
			}
			switch par & 0xff {
			case pInt64:
				data.params[key], err = strconv.ParseInt(val, 10, 64)
				if err != nil {
					logger.LogInfo(consts.StrToIntError, val)
				}
			case pHex:
				bin, err := hex.DecodeString(val)
				if err != nil {
					logger.LogError(consts.RouteError, err)
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
			logger.LogError(consts.RouteError, err)
			errorAPI(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(jsonResult)
	})
}
