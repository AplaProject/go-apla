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
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"

	"github.com/gorilla/schema"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	KeyID         int64
	EcosystemID   int64
	EcosystemName string
	RoleID        int64
	IsMobile      string
	IsVDE         bool
}

// Prefix returns prefix of ecosystem
func (c *Client) Prefix() (prefix string) {
	prefix = converter.Int64ToStr(c.EcosystemID)
	if c.IsVDE {
		prefix += `_vde`
	}
	return
}

// type forSign struct {
// 	Time    string `json:"time"`
// 	ForSign string `json:"forsign"`
// }

// // DefaultHandler is a common handle function for api requests
// func DefaultHandler(method, pattern string, params map[string]int, handlers ...apiHandle) hr.Handle {
// 	return hr.Handle(func(w http.ResponseWriter, r *http.Request, ps hr.Params) {
// 		counterName := statsd.APIRouteCounterName(method, pattern)
// 		statsd.Client.Inc(counterName+statsd.Count, 1, 1.0)
// 		startTime := time.Now()
// 		var (
// 			err  error
// 			data = &apiData{ecosystemId: 1}
// 		)

// 		// TODO: перенесено в LoggerMiddleware
// 		requestLogger := log.WithFields(log.Fields{"headers": r.Header, "path": r.URL.Path, "protocol": r.Proto, "remote": r.RemoteAddr})
// 		requestLogger.Info("received http request")

// 		defer func() {
// 			endTime := time.Now()
// 			statsd.Client.TimingDuration(counterName+statsd.Time, endTime.Sub(startTime), 1.0)
// 			if r := recover(); r != nil {
// 				requestLogger.WithFields(log.Fields{"type": consts.PanicRecoveredError, "error": r, "stack": string(debug.Stack())}).Error("panic recovered error")
// 				fmt.Println("API Recovered", fmt.Sprintf("%s: %s", r, debug.Stack()))
// 				errorAPI(w, `E_RECOVERED`, http.StatusInternalServerError)
// 			}
// 		}()

// 		w.Header().Set("Access-Control-Allow-Origin", "*")
// 		w.Header().Set("Content-Type", "application/json; charset=utf-8")

// 		data.params = make(map[string]interface{})
// 		for _, par := range ps {
// 			data.params[par.Key] = par.Value
// 		}

// 		ihandlers := append([]apiHandle{
// 			fillToken,
// 			fillParams(params),
// 		}, handlers...)

// 		for _, handler := range ihandlers {
// 			if handler(w, r, data, requestLogger) != nil {
// 				return
// 			}
// 		}

// 		jsonResult, err := json.Marshal(data.result)
// 		if err != nil {
// 			requestLogger.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marhsalling http response to json")
// 			errorAPI(w, err, http.StatusInternalServerError)
// 			return
// 		}

// 		w.Write(jsonResult)
// 	})
// }

func checkEcosystem(w http.ResponseWriter, r *http.Request) (ecosystemID int64, prefix string, ok bool) {
	client := getClient(r)
	ecosystemID = client.EcosystemID
	paramEcosystemID := converter.StrToInt64(r.FormValue("ecosystem"))
	if paramEcosystemID > 0 {
		logger := getLogger(r)
		count, err := model.GetNextID(nil, "1_ecosystems")
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting next id ecosystems")
			errorResponse(w, err, http.StatusInternalServerError)
			return
		}
		if paramEcosystemID >= count {
			logger.WithFields(log.Fields{"state_id": ecosystemID, "count": count, "type": consts.ParameterExceeded}).Error("state_id is larger then max count")
			errorResponse(w, errEcosystem, http.StatusBadRequest, ecosystemID)
			return
		}
		ecosystemID = paramEcosystemID
	}
	prefix = converter.Int64ToStr(ecosystemID)
	if client.IsVDE {
		prefix += `_vde`
	}
	return ecosystemID, prefix, true
}

type Form struct{}

func (f *Form) Validate(w http.ResponseWriter, r *http.Request) bool {
	return true
}

type FormValidater interface {
	Validate(w http.ResponseWriter, r *http.Request) bool
}

func ParseForm(w http.ResponseWriter, r *http.Request, f FormValidater) bool {
	r.ParseForm()
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	if err := decoder.Decode(f, r.Form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return false
	}
	return f.Validate(w, r)
}

type hexValue struct {
	value []byte
}

func (hv hexValue) Value() []byte {
	return hv.value
}

func (hv *hexValue) UnmarshalText(v []byte) (err error) {
	hv.value, err = hex.DecodeString(string(v))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "value": string(v), "error": err}).Error("decoding from hex")
	}
	return
}

type ecosystemForm struct {
	EcosystemID     int64  `schema:"ecosystem"`
	EcosystemPrefix string `schema:"-"`
}

func (f *ecosystemForm) Validate(w http.ResponseWriter, r *http.Request) bool {
	return f.ValidateEcosystem(w, r)
}

func (f *ecosystemForm) ValidateEcosystem(w http.ResponseWriter, r *http.Request) bool {
	client := getClient(r)
	logger := getLogger(r)

	if f.EcosystemID > 0 {
		count, err := model.GetNextID(nil, "1_ecosystems")
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting next id of ecosystems")
			errorResponse(w, err, http.StatusInternalServerError)
			return false
		}
		if f.EcosystemID >= count {
			logger.WithFields(log.Fields{"state_id": f.EcosystemID, "count": count, "type": consts.ParameterExceeded}).Error("ecosystem is larger then max count")
			errorResponse(w, errEcosystem, http.StatusBadRequest, f.EcosystemID)
			return false
		}
	} else {
		f.EcosystemID = client.EcosystemID
	}

	f.EcosystemPrefix = converter.Int64ToStr(f.EcosystemID)
	if client.IsVDE {
		f.EcosystemPrefix += `_vde`
	}

	return true
}
