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

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

type paramResult struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Value      string `json:"value"`
	Conditions string `json:"conditions"`
}

type ecosystemParamsResult struct {
	List []paramResult `json:"list"`
}

func getEcosystemParamsHandler(w http.ResponseWriter, r *http.Request) {
	form := &appParamsForm{}
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	logger := getLogger(r)

	sp := &model.StateParameter{}
	sp.SetTablePrefix(form.EcosystemPrefix)
	list, err := sp.GetAllStateParameters()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting all state parameters")
	}

	result := &ecosystemParamsResult{
		List: make([]paramResult, 0),
	}

	acceptNames := form.AcceptNames()
	for _, item := range list {
		if len(acceptNames) > 0 && !acceptNames[item.Name] {
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
