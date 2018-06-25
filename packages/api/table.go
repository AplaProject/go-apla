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
	"encoding/json"
	"net/http"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
	"github.com/gorilla/mux"
)

type columnInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Perm string `json:"perm"`
}

type tableResult struct {
	Name       string       `json:"name"`
	Insert     string       `json:"insert"`
	NewColumn  string       `json:"new_column"`
	Update     string       `json:"update"`
	Read       string       `json:"read,omitempty"`
	Filter     string       `json:"filter,omitempty"`
	Conditions string       `json:"conditions"`
	AppID      string       `json:"app_id"`
	Columns    []columnInfo `json:"columns"`
}

func tableHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	logger := getLogger(r)
	client := getClient(r)

	prefix := client.Prefix()

	table := &model.Table{}
	table.SetTablePrefix(prefix)
	_, err := table.Get(nil, strings.ToLower(params[keyName]))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting table")
		errorResponse(w, err)
		return
	}

	if len(table.Name) == 0 {
		errorResponse(w, errTableNotFound.Errorf(params[keyName]))
		return
	}

	var cols map[string]string
	err = json.Unmarshal([]byte(table.Columns), &cols)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("Unmarshalling table columns to json")
		errorResponse(w, err)
		return
	}
	columns := make([]columnInfo, 0)
	for key, value := range cols {
		colType, err := model.GetColumnType(prefix+`_`+params[keyName], key)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting column type from db")
			errorResponse(w, err)
			return
		}
		columns = append(columns, columnInfo{
			Name: key,
			Perm: value,
			Type: colType,
		})
	}

	jsonResponse(w, &tableResult{
		Name:       table.Name,
		Insert:     table.Permissions.Insert,
		NewColumn:  table.Permissions.NewColumn,
		Update:     table.Permissions.Update,
		Read:       table.Permissions.Read,
		Filter:     table.Permissions.Filter,
		Conditions: table.Conditions,
		AppID:      converter.Int64ToStr(table.AppID),
		Columns:    columns,
	})
}
