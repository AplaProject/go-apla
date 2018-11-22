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

type appParamsResult struct {
	App  string        `json:"app_id"`
	List []paramResult `json:"list"`
}

type appParamsForm struct {
	ecosystemForm
	paramsForm
}

func (f *appParamsForm) Validate(r *http.Request) error {
	return f.ecosystemForm.Validate(r)
}

func getAppParamsHandler(w http.ResponseWriter, r *http.Request) {
	form := &appParamsForm{}
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	logger := getLogger(r)

	ap := &model.AppParam{}
	ap.SetTablePrefix(form.EcosystemPrefix)

	list, err := ap.GetAllAppParameters(converter.StrToInt64(params["appID"]))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting all app parameters")
	}

	result := &appParamsResult{
		App:  params["appID"],
		List: make([]paramResult, 0),
	}

	acceptNames := form.AcceptNames()
	for _, item := range list {
		if !acceptNames[item.Name] {
			continue
		}
		result.List = append(result.List, paramResult{
			ID:         converter.Int64ToStr(item.ID),
			Name:       item.Name,
			Value:      item.Value,
			Conditions: item.Conditions,
		})
	}

	jsonResponse(w, result)
}
