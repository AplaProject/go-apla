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

	"github.com/EGaaS/go-egaas-mvp/packages/consts"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"
)

type columnItem struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Perm string `json:"perm"`
}

type tableResult struct {
	Name          string       `json:"name"`
	PermInsert    string       `json:"insert"`
	PermNewColumn string       `json:"new_column"`
	PermChange    string       `json:"general_update"`
	Columns       []columnItem `json:"columns"`
}

type tableItem struct {
	Name string `json:"name"`
}

type tableListResult struct {
	Count string      `json:"count"`
	List  []tableItem `json:"list"`
}

func getTable(w http.ResponseWriter, r *http.Request, data *apiData) error {
	logger.LogDebug(consts.FuncStarted, "")
	prefix := getPrefix(data)
	tableName := prefix + `_` + data.params[`name`].(string)
	tablePermission, err := model.GetMap(`SELECT data.* FROM "`+prefix+`_tables", jsonb_each_text(columns_and_permissions) as data WHERE name = ?`,
		"key", "value", tableName)
	if err != nil {
		logger.LogError(consts.DBError, err)
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	columnsAndPermissions, err := model.GetMap(`SELECT data.* FROM "`+prefix+`_tables", jsonb_each_text(columns_and_permissions->'update') as data WHERE name = ?`,
		"key", "value", tableName)
	if err != nil {
		logger.LogError(consts.DBError, err)
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	columns := make([]columnItem, 0)
	for key, value := range columnsAndPermissions {
		columns = append(columns, columnItem{Name: key, Perm: value, Type: model.GetColumnType(tableName, key)})
	}
	data.result = &tableResult{Name: tableName, PermInsert: tablePermission[`insert`],
		PermNewColumn: tablePermission[`new_column`], PermChange: tablePermission[`general_update`],
		Columns: columns}
	return nil
}

func txPreNewTable(w http.ResponseWriter, r *http.Request, data *apiData) error {
	logger.LogDebug(consts.FuncStarted, "")
	v := tx.NewTable{
		Header:  getSignHeader(`NewTable`, data),
		Global:  converter.Int64ToStr(data.params[`global`].(int64)),
		Name:    data.params[`name`].(string),
		Columns: data.params[`columns`].(string),
	}
	data.result = &forSign{Time: converter.Int64ToStr(v.Time), ForSign: v.ForSign()}
	return nil
}

func txNewTable(w http.ResponseWriter, r *http.Request, data *apiData) error {
	logger.LogDebug(consts.FuncStarted, "")
	header, err := getHeader(`NewTable`, data)
	if err != nil {
		logger.LogError(consts.GetHeaderError, err)
		return errorAPI(w, err.Error(), http.StatusBadRequest)
	}
	toSerialize := tx.NewTable{
		Header:  header,
		Global:  converter.Int64ToStr(data.params[`global`].(int64)),
		Name:    data.params[`name`].(string),
		Columns: data.params[`columns`].(string),
	}
	hash, err := sendEmbeddedTx(header.Type, header.UserID, toSerialize)
	if err != nil {
		logger.LogError(consts.SendEmbeddedTxError, err)
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	data.result = hash
	return nil
}

func txPreEditTable(w http.ResponseWriter, r *http.Request, data *apiData) error {
	logger.LogDebug(consts.FuncStarted, "")
	v := tx.EditTable{
		Header:        getSignHeader(`EditTable`, data),
		Name:          data.params[`name`].(string),
		GeneralUpdate: data.params[`general_update`].(string),
		Insert:        data.params[`insert`].(string),
		NewColumn:     data.params[`new_column`].(string),
	}
	data.result = &forSign{Time: converter.Int64ToStr(v.Time), ForSign: v.ForSign()}
	return nil
}

func txEditTable(w http.ResponseWriter, r *http.Request, data *apiData) error {
	logger.LogDebug(consts.FuncStarted, "")
	header, err := getHeader(`EditTable`, data)
	if err != nil {
		logger.LogError(consts.GetHeaderError, err)
		return errorAPI(w, err.Error(), http.StatusBadRequest)
	}
	toSerialize := tx.EditTable{
		Header:        header,
		Name:          data.params[`name`].(string),
		GeneralUpdate: data.params[`general_update`].(string),
		Insert:        data.params[`insert`].(string),
		NewColumn:     data.params[`new_column`].(string),
	}
	hash, err := sendEmbeddedTx(header.Type, header.UserID, toSerialize)
	if err != nil {
		logger.LogError(consts.SendEmbeddedTxError, err)
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	data.result = hash
	return nil
}

func tableList(w http.ResponseWriter, r *http.Request, data *apiData) error {
	logger.LogDebug(consts.FuncStarted, "")
	limit := int(data.params[`limit`].(int64))
	if limit == 0 {
		limit = 25
	} else if limit < 0 {
		limit = -1
	}
	outList := make([]tableItem, 0)
	count, err := model.Single(`SELECT count(*) FROM "` + getPrefix(data) + `_tables"`).String()
	if err != nil {
		logger.LogError(consts.DBError, err)
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}

	list, err := model.GetAll(`SELECT name FROM "`+getPrefix(data)+`_tables" order by name`+
		fmt.Sprintf(` offset %d `, data.params[`offset`].(int64)), limit)
	if err != nil {
		logger.LogError(consts.DBError, err)
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}

	for _, val := range list {
		outList = append(outList, tableItem{Name: val[`name`]})
	}
	data.result = &tableListResult{Count: count, List: outList}
	return nil
}

func txPreEditColumn(w http.ResponseWriter, r *http.Request, data *apiData) error {
	logger.LogDebug(consts.FuncStarted, "")
	v := tx.EditColumn{
		Header:      getSignHeader(`EditColumn`, data),
		TableName:   data.params[`table`].(string),
		ColumnName:  data.params[`name`].(string),
		Permissions: data.params[`permissions`].(string),
	}
	data.result = &forSign{Time: converter.Int64ToStr(v.Time), ForSign: v.ForSign()}
	return nil
}

func txEditColumn(w http.ResponseWriter, r *http.Request, data *apiData) error {
	logger.LogDebug(consts.FuncStarted, "")
	header, err := getHeader(`EditColumn`, data)
	if err != nil {
		logger.LogError(consts.GetHeaderError, err)
		return errorAPI(w, err.Error(), http.StatusBadRequest)
	}
	toSerialize := tx.EditColumn{
		Header:      header,
		TableName:   data.params[`table`].(string),
		ColumnName:  data.params[`name`].(string),
		Permissions: data.params[`permissions`].(string),
	}
	hash, err := sendEmbeddedTx(header.Type, header.UserID, toSerialize)
	if err != nil {
		logger.LogError(consts.SendEmbeddedTxError, err)
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	data.result = hash
	return nil
}

func txPreNewColumn(w http.ResponseWriter, r *http.Request, data *apiData) error {
	logger.LogDebug(consts.FuncStarted, "")
	v := tx.NewColumn{
		Header:      getSignHeader(`NewColumn`, data),
		TableName:   data.params[`table`].(string),
		ColumnName:  data.params[`name`].(string),
		ColumnType:  data.params[`type`].(string),
		Permissions: data.params[`permissions`].(string),
		Index:       converter.Int64ToStr(data.params[`index`].(int64)),
	}
	data.result = &forSign{Time: converter.Int64ToStr(v.Time), ForSign: v.ForSign()}
	return nil
}

func txNewColumn(w http.ResponseWriter, r *http.Request, data *apiData) error {
	logger.LogDebug(consts.FuncStarted, "")
	header, err := getHeader(`NewColumn`, data)
	if err != nil {
		logger.LogError(consts.GetHeaderError, err)
		return errorAPI(w, err.Error(), http.StatusBadRequest)
	}
	toSerialize := tx.NewColumn{
		Header:      header,
		TableName:   data.params[`table`].(string),
		ColumnName:  data.params[`name`].(string),
		ColumnType:  data.params[`type`].(string),
		Permissions: data.params[`permissions`].(string),
		Index:       converter.Int64ToStr(data.params[`index`].(int64)),
	}
	hash, err := sendEmbeddedTx(header.Type, header.UserID, toSerialize)
	if err != nil {
		logger.LogDebug(consts.SendEmbeddedTxError, err)
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	data.result = hash
	return nil
}
