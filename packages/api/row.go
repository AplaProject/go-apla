// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package api

import (
	"net/http"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type rowResult struct {
	Value map[string]string `json:"value"`
}

type rowForm struct {
	Columns string `schema:"columns"`
}

func (f *rowForm) Validate(r *http.Request) error {
	if len(f.Columns) > 0 {
		f.Columns = converter.EscapeName(f.Columns)
	}
	return nil
}

func getRowHandler(w http.ResponseWriter, r *http.Request) {
	form := &rowForm{}
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	client := getClient(r)
	logger := getLogger(r)

	q := model.GetDB(nil).Limit(1)

	var (
		err   error
		table string
	)
	table, form.Columns, err = checkAccess(params["name"], form.Columns, client)
	if err != nil {
		errorResponse(w, err)
		return
	}
	col := `id`
	if len(params["column"]) > 0 {
		col = converter.Sanitize(params["column"], `-`)
	}
	if converter.FirstEcosystemTables[params["name"]] {
		q = q.Table(table).Where(col+" = ? and ecosystem = ?", params["id"], client.EcosystemID)
	} else {
		q = q.Table(table).Where(col+" = ?", params["id"])
	}

	if len(form.Columns) > 0 {
		q = q.Select(form.Columns)
	}

	rows, err := q.Rows()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "table": table}).Error("Getting rows from table")
		errorResponse(w, errQuery)
		return
	}

	result, err := model.GetResult(rows)
	if err != nil {
		errorResponse(w, err)
		return
	}

	if len(result) == 0 {
		errorResponse(w, errNotFound)
		return
	}

	jsonResponse(w, &rowResult{
		Value: result[0],
	})
}
