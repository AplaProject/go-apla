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
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type listResult struct {
	Count string              `json:"count"`
	List  []map[string]string `json:"list"`
}

type listForm struct {
	paginatorForm
	Columns string `schema:"columns"`
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	form := &listForm{}
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err)
		return
	}

	params := mux.Vars(r)
	client := getClient(r)
	logger := getLogger(r)

	// TODO: перенести в модели
	table := converter.EscapeName(fmt.Sprintf("%s_%s", client.Prefix(), params[keyName]))
	cols := `*`
	if len(form.Columns) > 0 {
		cols = `id,` + converter.EscapeName(form.Columns)
	}

	count, err := model.GetRecordsCountTx(nil, strings.Trim(table, `"`))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "table": table}).Error("Getting table records count")
		errorResponse(w, errTableNotFound.Errorf(params[keyName]))
		return
	}

	list, err := model.GetAll(`select `+cols+` from `+table+` order by id desc`+
		fmt.Sprintf(` offset %d `, form.Offset), form.Limit)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "table": table}).Error("Getting rows from table")
		errorResponse(w, err)
		return
	}

	jsonResponse(w, &listResult{
		Count: converter.Int64ToStr(count),
		List:  list,
	})
}
