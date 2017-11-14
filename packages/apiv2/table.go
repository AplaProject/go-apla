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
	"encoding/json"
	"net/http"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
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
	Conditions string       `json:"conditions"`
	Columns    []columnInfo `json:"columns"`
}

func table(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	var result tableResult

	prefix := getPrefix(data)
	table := &model.Table{}
	table.SetTablePrefix(prefix)
	_, err = table.Get(data.params[`name`].(string))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting table")
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}

	if len(table.Name) > 0 {
		var perm map[string]string
		err := json.Unmarshal([]byte(table.Permissions), &perm)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("Unmarshalling table permissions to json")
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}
		var cols map[string]string
		err = json.Unmarshal([]byte(table.Columns), &cols)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("Unmarshalling table columns to json")
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}
		columns := make([]columnInfo, 0)
		for key, value := range cols {
			colType, err := model.GetColumnType(prefix+`_`+data.params[`name`].(string), key)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting column type from db")
				return errorAPI(w, err.Error(), http.StatusInternalServerError)
			}
			columns = append(columns, columnInfo{Name: key, Perm: value,
				Type: colType})
		}
		result = tableResult{
			Name:       table.Name,
			Insert:     perm[`insert`],
			NewColumn:  perm[`new_column`],
			Update:     perm[`update`],
			Conditions: table.Conditions,
			Columns:    columns,
		}
	} else {
		return errorAPI(w, `E_TABLENOTFOUND`, http.StatusBadRequest, data.params[`name`].(string))
	}
	data.result = &result
	return
}
