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
	"fmt"
	"net/http"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
)

type tableInfo struct {
	Name string `json:"name"`
}

type tablesResult struct {
	Count int64       `json:"count"`
	List  []tableInfo `json:"list"`
}

func tables(w http.ResponseWriter, r *http.Request, data *apiData) (err error) {
	var (
		result tablesResult
		limit  int
	)

	table := converter.Int64ToStr(data.state) + `_tables`

	count, err := model.Single(`select count(*) from "` + table + `"`).Int64()
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	if data.params[`limit`].(int64) > 0 {
		limit = int(data.params[`limit`].(int64))
	} else {
		limit = 25
	}
	list, err := model.GetAll(`select name from "`+table+`" order by name`+
		fmt.Sprintf(` offset %d `, data.params[`offset`].(int64)), limit)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}

	result = tablesResult{
		Count: count, List: make([]tableInfo, len(list)),
	}
	for i, item := range list {
		result.List[i].Name = item[`name`]
	}
	data.result = &result
	return
}
