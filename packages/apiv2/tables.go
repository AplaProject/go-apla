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

type tableInfo struct {
	Name string `json:"name"`
}

type tablesResult struct {
	Count int         `json:"count"`
	List  []tableInfo `json:"list"`
}

func tables(w http.ResponseWriter, r *http.Request, data *apiData) (err error) {
	var result tablesResult

	result = tablesResult{
		Count: 22, List: []tableInfo{{Name: `citizens`}, {Name: `contracts`}},
	}
	data.result = &result
	return
}
