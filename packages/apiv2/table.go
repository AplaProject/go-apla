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

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
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

func table(w http.ResponseWriter, r *http.Request, data *apiData) (err error) {
	var result tableResult

	row, err := model.GetOneRow(`select * from "`+converter.Int64ToStr(data.state)+`_tables" where name=?`,
		data.params[`name`].(string)).String()
	if len(row[`name`]) > 0 {
		var perm map[string]string
		err := json.Unmarshal([]byte(row[`permissions`]), &perm)
		if err != nil {
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}
		var cols map[string]string
		err = json.Unmarshal([]byte(row[`columns`]), &cols)
		if err != nil {
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}
		columns := make([]columnInfo, 0)
		for key, value := range cols {
			colType, err := model.GetColumnType(converter.Int64ToStr(data.state)+`_`+data.params[`name`].(string), key)
			if err != nil {
				return errorAPI(w, err.Error(), http.StatusInternalServerError)
			}
			columns = append(columns, columnInfo{Name: key, Perm: value,
				Type: colType})
		}
		result = tableResult{
			Name:       row[`name`],
			Insert:     perm[`insert`],
			NewColumn:  perm[`new_column`],
			Update:     perm[`update`],
			Conditions: row[`conditions`],
			Columns:    columns,
		}
	}
	data.result = &result
	return
}
