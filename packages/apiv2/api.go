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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"runtime/debug"
	"strings"
	"time"

	"github.com/AplaProject/go-apla/packages/config"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/AplaProject/go-apla/packages/utils/tx"
	"github.com/dgrijalva/jwt-go"
	hr "github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
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

type apiHandle func(http.ResponseWriter, *http.Request, *apiData, *log.Entry) error

var (
	installed bool
)

func errorAPI(w http.ResponseWriter, err interface{}, code int, params ...interface{}) error {
	var (
		msg, errCode, errParams string
	)

	switch v := err.(type) {
	case string:
		errCode = v
		if val, ok := errors[v]; ok {
			if len(params) > 0 {
				list := make([]string, 0)
				msg = fmt.Sprintf(val, params...)
				for _, item := range params {
					list = append(list, fmt.Sprintf(`"%v"`, item))
				}
				errParams = fmt.Sprintf(`, "params": [%s]`, strings.Join(list, `,`))
			} else {
				msg = val
			}
		} else {
			msg = v
		}
	case interface{}:
		errCode = `E_SERVER`
		if reflect.TypeOf(v).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			msg = v.(error).Error()
		}
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintln(w, fmt.Sprintf(`{"error": %q, "msg": %q %s}`, errCode, msg, errParams))
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
		log.WithFields(log.Fields{"type": consts.EmptyObject, "params": data.params}).Error("signature is empty")
		return tx.Header{}, fmt.Errorf("signature is empty")
	}
	return tx.Header{Type: int(utils.TypeInt(txName)), Time: converter.StrToInt64(data.params[`time`].(string)),
		UserID: data.wallet, StateID: data.state, PublicKey: publicKey,
		BinSignatures: converter.EncodeLengthPlusData(signature)}, nil
}

// DefaultHandler is a common handle function for api requests
func DefaultHandler(params map[string]int, handlers ...apiHandle) hr.Handle {
	return hr.Handle(func(w http.ResponseWriter, r *http.Request, ps hr.Params) {
		var (
			err  error
			data apiData
		)
		requestLogger := log.WithFields(log.Fields{"headers": r.Header, "path": r.URL.Path, "protocol": r.Proto, "remote": r.RemoteAddr})
		requestLogger.Info("received http request")
		defer func() {
			if r := recover(); r != nil {
				requestLogger.WithFields(log.Fields{"type": consts.PanicRecoveredError, "error": r, "stack": string(debug.Stack())}).Error("panic recovered error")
				fmt.Println("API Recovered", fmt.Sprintf("%s: %s", r, debug.Stack()))
				errorAPI(w, `E_RECOVERED`, http.StatusInternalServerError)
			}
		}()
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if !installed && r.URL.Path != `/api/v2/install` {
			if model.DBConn == nil && !config.IsExist() {
				errorAPI(w, `E_NOTINSTALLED`, http.StatusInternalServerError)
				return
			}
			installed = true
		}
		token, err := jwtToken(r)
		if err != nil {
			requestLogger.WithFields(log.Fields{"type": consts.JWTError, "params": params, "error": err}).Error("starting session")
			errmsg := err.Error()
			expired := `token is expired by`
			if strings.HasPrefix(errmsg, expired) {
				errorAPI(w, `E_TOKENEXPIRED`, http.StatusUnauthorized, errmsg[len(expired):])
				return
			}
			errorAPI(w, err, http.StatusBadRequest)
			return
		}
		data.token = token
		if token != nil && token.Valid {
			if claims, ok := token.Claims.(*JWTClaims); ok && len(claims.Wallet) > 0 {
				data.state = converter.StrToInt64(claims.State)
				data.wallet = converter.StrToInt64(claims.Wallet)
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
				requestLogger.WithFields(log.Fields{"type": consts.RouteError, "error": fmt.Sprintf("undefined val %s", key)}).Error("undefined val")
				errorAPI(w, `E_UNDEFINEVAL`, http.StatusBadRequest, key)
				return
			}
			switch par & 0xff {
			case pInt64:
				data.params[key] = converter.StrToInt64(val)
			case pHex:
				bin, err := hex.DecodeString(val)
				if err != nil {
					requestLogger.WithFields(log.Fields{"type": consts.ConvertionError, "value": val, "error": err}).Error("decoding http parameter from hex")
					errorAPI(w, err, http.StatusBadRequest)
					return
				}
				data.params[key] = bin
			case pString:
				data.params[key] = val
			}
		}
		for _, handler := range handlers {
			if handler(w, r, &data, requestLogger) != nil {
				return
			}
		}
		jsonResult, err := json.Marshal(data.result)
		if err != nil {
			requestLogger.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marhsalling http response to json")
			errorAPI(w, err, http.StatusInternalServerError)
			return
		}
		w.Write(jsonResult)
	})
}

func checkEcosystem(w http.ResponseWriter, data *apiData, logger *log.Entry) (int64, error) {
	state := data.state
	if data.params[`ecosystem`].(int64) > 0 {
		state = data.params[`ecosystem`].(int64)
		count, err := model.GetNextID(nil, `system_states`)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting next id system states")
			return 0, errorAPI(w, err, http.StatusBadRequest)
		}
		if state >= count {
			logger.WithFields(log.Fields{"state_id": state, "count": count, "type": consts.ParameterExceeded}).Error("state_id is larger then max count")
			return 0, errorAPI(w, `E_ECOSYSTEM`, http.StatusBadRequest, state)
		}
	}
	return state, nil
}
