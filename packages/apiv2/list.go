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
	"strings"

	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
)

type listResult struct {
	Count string              `json:"count"`
	List  []map[string]string `json:"list"`
}

func list(w http.ResponseWriter, r *http.Request, data *apiData) (err error) {
	var limit int

	table := converter.EscapeName(converter.Int64ToStr(data.state) + `_` + data.params[`name`].(string))
	cols := `*`
	if len(data.params[`columns`].(string)) > 0 {
		cols = `id,` + converter.EscapeName(data.params[`columns`].(string))
	}

	count, err := model.GetNextID(strings.Trim(table, `"`))
	if err != nil {
		return errorAPI(w, `E_TABLENOTFOUND`, http.StatusBadRequest, data.params[`name`].(string))
	}

	if data.params[`limit`].(int64) > 0 {
		limit = int(data.params[`limit`].(int64))
	} else {
		limit = 25
	}
	list, err := model.GetAll(`select `+cols+` from `+table+` order by id desc`+
		fmt.Sprintf(` offset %d `, data.params[`offset`].(int64)), limit)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	data.result = &listResult{
		Count: converter.Int64ToStr(count - 1), List: list,
	}
	return
}
