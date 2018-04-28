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
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	keyAppID = "id"
	keyName  = "name"
)

func appParamHandler(w http.ResponseWriter, r *http.Request) {
	form := &ecosystemForm{}
	if ok := ParseForm(w, r, form); !ok {
		return
	}

	logger := getLogger(r)
	params := mux.Vars(r)

	ap := &model.AppParam{}
	ap.SetTablePrefix(form.EcosystemPrefix)
	found, err := ap.Get(nil, converter.StrToInt64(params[keyAppID]), params[keyName])
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting app parameter by name")
		errorResponse(w, err, http.StatusInternalServerError)
		return
	}
	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound, "key": params[keyName]}).Error("app parameter not found")
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	jsonResponse(w, &paramResult{
		ID:         converter.Int64ToStr(ap.ID),
		Name:       ap.Name,
		Value:      ap.Value,
		Conditions: ap.Conditions,
	})
}
