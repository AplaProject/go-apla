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
	"fmt"
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
)

type listResult struct {
	Count string              `json:"count"`
	List  []map[string]string `json:"list"`
}

func list(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	var limit int

	table := converter.ParseTable(data.params[`name`].(string), data.ecosystemId)
	cols := `*`
	if len(data.params[`columns`].(string)) > 0 {
		cols = `id,` + converter.EscapeName(data.params[`columns`].(string))
	}

	count, err := model.GetRecordsCountTx(nil, table)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "table": table}).Error("Getting table records count")
		return errorAPI(w, `E_TABLENOTFOUND`, http.StatusBadRequest, data.params[`name`].(string))
	}

	if data.params[`limit`].(int64) > 0 {
		limit = int(data.params[`limit`].(int64))
	} else {
		limit = 25
	}
	var query string
	query = fmt.Sprintf(`select %s from "%s" order by id desc offset %d `, cols, table,
		data.params[`offset`].(int64))
	list, err := model.GetAll(query, limit)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "table": table}).Error("Getting rows from table")
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	data.result = &listResult{
		Count: converter.Int64ToStr(count), List: list,
	}
	return
}
