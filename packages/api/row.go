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

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const keyColumns = "columns"
const keyID = "id"

type rowResult struct {
	Value map[string]string `json:"value"`
}

type rowForm struct {
	Columns string `schema:"columns"`
}

func rowHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	client := getClient(r)
	logger := getLogger(r)

	tableName := client.Prefix() + "_" + params[keyName]
	row, err := model.GetRowByID(tableName, r.FormValue(keyColumns), converter.StrToInt64(params[keyID]))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "table": params[keyName], "id": params[keyID]}).Error("getting one row")
		errorResponse(w, errQuery)
		return
	}

	jsonResponse(w, &rowResult{
		Value: row,
	})
}
