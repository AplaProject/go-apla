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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/types"
	"github.com/AplaProject/go-apla/packages/utils/tx"
	"github.com/gorilla/mux"
)

type dbFindResult struct {
	List []interface{}
}

type dbfindForm struct {
	ID      int64  `schema:"id"`
	Order   string `schema:"order"`
	Columns string `schema:"columns"`
	paginatorForm
}

func (f *dbfindForm) Validate(r *http.Request) error {
	if err := f.paginatorForm.Validate(r); err != nil {
		return err
	}

	if len(f.Columns) > 0 {
		f.Columns = converter.EscapeName(f.Columns)
	}

	return nil
}

func getDbFindHandler(w http.ResponseWriter, r *http.Request) {
	form := &dbfindForm{}
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	client := getClient(r)
	tableName := strings.ToLower(params["table"])
	sc := smart.SmartContract{
		OBS: conf.Config.IsSupportingOBS(),
		VM:  smart.GetVM(),
		TxSmart: tx.SmartContract{
			Header: tx.Header{
				EcosystemID: client.EcosystemID,
				KeyID:       client.KeyID,
				NetworkID:   conf.Config.NetworkID,
			},
		},
	}

	// Check table existence
	prefix := client.Prefix()
	table := &model.Table{}
	table.SetTablePrefix(prefix)
	if found, err := table.Get(nil, tableName); !found || nil != err {
		errorResponse(w, errTableNotFound.Errorf(tableName))
		return
	}

	// Unmarshall where clause if there is any
	var formWhere map[string]interface{}
	cols, err := table.GetColumns(nil, tableName, "")
	fmt.Printf("%+v %+v\n", cols, err)

	cols, err = table.GetColumns(nil, tableName, form.Columns)
	fmt.Printf("%+v %+v\n", cols, err)

	if whereValue := r.FormValue("where"); 0 < len(whereValue) {
		if err := json.Unmarshal([]byte(r.FormValue("where")), &formWhere); err != nil {
			errorResponse(w, err, http.StatusBadRequest)
			return
		}
	}
	whereClause := types.LoadMap(formWhere)

	// Perform the actual request
	_, ret, err := smart.DBSelect(&sc, tableName, form.Columns, form.ID, form.Order, form.Offset, form.Limit, whereClause)
	if err != nil {
		errorResponse(w, err)
		return
	}

	result := new(dbFindResult)
	result.List = ret
	jsonResponse(w, result)
}
