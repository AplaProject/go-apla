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

	log "github.com/sirupsen/logrus"
)

type rowResult struct {
	Value map[string]string `json:"value"`
}

func getTableRowByID(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	data.params["column"] = "id"
	return getTableRow(w, r, data, logger)
}

func getTableRow(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	columns := data.ParamString("columns")
	if len(columns) > 0 {
		columns = converter.EscapeName(columns)
	} else {
		columns = "*"
	}
	tableSuffix := data.ParamString("name")
	table := converter.EscapeName(getPrefix(data) + `_` + tableSuffix)
	column := converter.EscapeName(data.ParamString("column"))
	value := data.ParamString("value")
	row, err := model.GetOneRow(`SELECT `+columns+` FROM `+table+` WHERE `+column+` = ?`, value).String()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "table": tableSuffix, "column": column, "value": value}).Error("getting one row")
		return errorAPI(w, `E_QUERY`, http.StatusInternalServerError)
	}

	data.result = &rowResult{Value: row}
	return
}
