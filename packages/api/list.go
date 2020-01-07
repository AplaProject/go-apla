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
	"fmt"
	"regexp"
	"errors"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/utils/tx"
	"github.com/AplaProject/go-apla/packages/converter"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type listResult struct {
	Count string              `json:"count"`
	List  []map[string]string `json:"list"`
}

type modelListResult struct {
	Count string	   `json:"count"`
	List  interface{}  `json:"list"`
}

type TableName struct {
	prefix int64
	suffix string
}

func ToTableName(full string) (tableName TableName, err error) {
	pat := fmt.Sprintf(`^(([0-9]+)_)?(.*)$`)
	re := regexp.MustCompile(pat)
	m := re.FindStringSubmatch(full)
	if m == nil {
		return tableName, errors.New(fmt.Sprintf("Full table name '%s' doesn't match to pattern '%s'", full, pat))
	}
	return TableName{prefix: converter.StrToInt64(m[2]), suffix: m[3]}, nil
}

type listForm struct {
	paginatorForm
	rowForm
}

func (f *listForm) Validate(r *http.Request) error {
	if err := f.paginatorForm.Validate(r); err != nil {
		return err
	}
	return f.rowForm.Validate(r)
}

func checkAccess(tableName, columns string, client *Client) (table string, cols string, err error) {
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
	table, _, cols, err = sc.CheckAccess(tableName, columns, client.EcosystemID)
	return
}

func getListHandler(w http.ResponseWriter, r *http.Request) {
	form := &listForm{}
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	client := getClient(r)
	logger := getLogger(r)

	var (
		err   error
		table string
	)
	table, form.Columns, err = checkAccess(params["name"], form.Columns, client)
	if err != nil {
		errorResponse(w, err)
		return
	}
	q := model.GetTableQuery(params["name"], client.EcosystemID)

	if len(form.Columns) > 0 {
		q = q.Select("id," + form.Columns)
	}

	tableName, err := ToTableName(params["name"])
	if err != nil {
		errorResponse(w, err)
	}
	order := "id ASC"
	switch tableName.suffix {
	case "keys":
		list := model.GetKeys(tableName.prefix, order, form.Limit, form.Offset)
		cnt, err := model.GetKeysCount()
		if err != nil {
		    errorResponse(w, err)
		}
		result := modelListResult{Count: converter.Int64ToStr(cnt), List: list}
		jsonResponse(w, result)
	default:
		result := new(listResult)
		err = q.Count(&result.Count).Error
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "table": table}).Error("Getting table records count")
			errorResponse(w, errTableNotFound.Errorf(table))
			return
		}

		rows, err := q.Order(order).Offset(form.Offset).Limit(form.Limit).Rows()
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "table": table}).Error("Getting rows from table")
			errorResponse(w, err)
			return
		}
		result.List, err = model.GetResult(rows)
		if err != nil {
			errorResponse(w, err)
			return
		}
		jsonResponse(w, result)
	}
}
