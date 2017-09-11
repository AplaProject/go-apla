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
	"net/http"
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

	result = tableResult{
		Name:       "mytable",
		Insert:     `ContractConditions("MainCondition")`,
		NewColumn:  `ContractConditions("MainCondition")`,
		Update:     `ContractConditions("MainCondition")`,
		Conditions: `ContractConditions("MainCondition")`,
		Columns: []columnInfo{{Name: "mynum", Type: "numbers", Perm: "ContractConditions(`MainCondition`)"},
			{Name: "mytext", Type: "text", Perm: "ContractConditions(`MainCondition`)"}},
	}
	data.result = &result
	return
}
