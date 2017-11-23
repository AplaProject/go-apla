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
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

type paramValue struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Value      string `json:"value"`
	Conditions string `json:"conditions"`
}

type ecosystemParamsResult struct {
	List []paramValue `json:"list"`
}

func ecosystemParams(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	var (
		result ecosystemParamsResult
		names  map[string]bool
	)
	_, prefix, err := checkEcosystem(w, data, logger)
	if err != nil {
		return err
	}
	sp := &model.StateParameter{}
	sp.SetTablePrefix(prefix)
	list, err := sp.GetAllStateParameters()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting all state parameters")
	}
	result.List = make([]paramValue, 0)
	if len(data.params[`names`].(string)) > 0 {
		names = make(map[string]bool)
		for _, item := range strings.Split(data.params[`names`].(string), `,`) {
			names[item] = true
		}
	}
	for _, item := range list {
		if names != nil && !names[item.Name] {
			continue
		}
		result.List = append(result.List, paramValue{ID: converter.Int64ToStr(item.ID),
			Name: item.Name, Value: item.Value, Conditions: item.Conditions})
	}
	data.result = &result
	return
}
