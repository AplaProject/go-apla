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
	//	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/astaxie/beego/session"
	hr "github.com/julienschmidt/httprouter"
	"github.com/op/go-logging"
)

type apiData struct {
	status int
	result interface{}
	params map[string]interface{}
	sess   session.SessionStore
}

const (
	pInt64 = iota
	pHex

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

func errConflict(w http.ResponseWriter, msg string) error {
	http.Error(w, msg, http.StatusConflict)
	return fmt.Errorf(msg)
}

func errBadRequest(w http.ResponseWriter, msg string) error {
	http.Error(w, msg, http.StatusBadRequest)
	return fmt.Errorf(msg)
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
				log.Error("API Recovered", r)
			}
		}()
		if apiSess == nil {
			errConflict(w, `Session is undefined`)
			return
		}

		data.sess, err = apiSess.SessionStart(w, r)
		if err != nil {
			errConflict(w, err.Error())
			return
		}
		defer data.sess.SessionRelease(w)

		// Getting and validating request parameters
		r.ParseForm()
		data.params = make(map[string]interface{})
		for key, par := range params {
			val := r.FormValue(key)
			if par&pOptional == 0 && len(val) == 0 {
				errBadRequest(w, fmt.Sprintf(`Value %s is undefined`, key))
				return
			}
			switch par & 0xff {
			case pInt64:
				data.params[key] = converter.StrToInt64(val)
			case pHex:
				bin, err := hex.DecodeString(val)
				if err != nil {
					errBadRequest(w, err.Error())
					return
				}
				data.params[key] = bin
			}
		}
		for _, handler := range handlers {
			if handler(w, r, &data) != nil {
				return
			}
		}
		jsonResult, err := json.Marshal(data.result)
		if err != nil {
			errConflict(w, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(jsonResult)
	})
}
