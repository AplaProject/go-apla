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

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/sql"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"
)

type menuResult struct {
	Name       string `json:"name"`
	Value      string `json:"value"`
	Conditions string `json:"conditions"`
}

type menuItem struct {
	Name string `json:"name"`
}

type menuListResult struct {
	Count string     `json:"count"`
	List  []menuItem `json:"list"`
}

func getMenu(w http.ResponseWriter, r *http.Request, data *apiData) error {

	dataMenu, err := sql.DB.OneRow(`SELECT * FROM "`+getPrefix(data)+`_menu" WHERE name = ?`,
		data.params[`name`].(string)).String()
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	data.result = &menuResult{Name: dataMenu["name"], Value: dataMenu["value"], Conditions: dataMenu["conditions"]}
	return nil
}

func txPreNewMenu(w http.ResponseWriter, r *http.Request, data *apiData) error {
	v := tx.NewMenu{
		Header:     getSignHeader(`NewMenu`, data),
		Global:     converter.Int64ToStr(data.params[`global`].(int64)),
		Name:       data.params[`name`].(string),
		Value:      data.params[`value`].(string),
		Conditions: data.params[`conditions`].(string),
	}
	data.result = &forSign{Time: converter.Int64ToStr(v.Time), ForSign: v.ForSign()}
	return nil
}

func txPreEditMenu(w http.ResponseWriter, r *http.Request, data *apiData) error {
	v := tx.EditMenu{
		Header:     getSignHeader(`EditMenu`, data),
		Global:     converter.Int64ToStr(data.params[`global`].(int64)),
		Name:       data.params[`name`].(string),
		Value:      data.params[`value`].(string),
		Conditions: data.params[`conditions`].(string),
	}
	data.result = &forSign{Time: converter.Int64ToStr(v.Time), ForSign: v.ForSign()}
	return nil
}

func txMenu(w http.ResponseWriter, r *http.Request, data *apiData) error {
	txName := `NewMenu`
	if r.Method == `PUT` {
		txName = `EditMenu`
	}
	header, err := getHeader(txName, data)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusBadRequest)
	}

	var toSerialize interface{}

	if txName == `EditMenu` {
		toSerialize = tx.EditMenu{
			Header:     header,
			Global:     converter.Int64ToStr(data.params[`global`].(int64)),
			Name:       data.params[`name`].(string),
			Value:      data.params[`value`].(string),
			Conditions: data.params[`conditions`].(string),
		}
	} else {
		toSerialize = tx.NewMenu{
			Header:     header,
			Global:     converter.Int64ToStr(data.params[`global`].(int64)),
			Name:       data.params[`name`].(string),
			Value:      data.params[`value`].(string),
			Conditions: data.params[`conditions`].(string),
		}
	}
	hash, err := sendEmbeddedTx(header.Type, header.UserID, toSerialize)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	data.result = hash
	return nil
}

func menuList(w http.ResponseWriter, r *http.Request, data *apiData) error {

	limit := int(data.params[`limit`].(int64))
	if limit == 0 {
		limit = 25
	} else if limit < 0 {
		limit = -1
	}
	outList := make([]menuItem, 0)
	count, err := sql.DB.Single(`SELECT count(*) FROM "` + getPrefix(data) + `_menu"`).String()
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}

	list, err := sql.DB.GetAll(`SELECT name FROM "`+getPrefix(data)+`_menu" order by name`+
		fmt.Sprintf(` offset %d `, data.params[`offset`].(int64)), limit)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}

	for _, val := range list {
		outList = append(outList, menuItem{Name: val[`name`]})
	}
	data.result = &menuListResult{Count: count, List: outList}
	return nil
}

func txPreAppendMenu(w http.ResponseWriter, r *http.Request, data *apiData) error {
	v := tx.AppendMenu{
		Header: getSignHeader(`AppendMenu`, data),
		Global: converter.Int64ToStr(data.params[`global`].(int64)),
		Name:   data.params[`name`].(string),
		Value:  data.params[`value`].(string),
	}
	data.result = &forSign{Time: converter.Int64ToStr(v.Time), ForSign: v.ForSign()}
	return nil
}

func txAppendMenu(w http.ResponseWriter, r *http.Request, data *apiData) error {
	header, err := getHeader(`AppendMenu`, data)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusBadRequest)
	}
	toSerialize := tx.AppendMenu{
		Header: header,
		Global: converter.Int64ToStr(data.params[`global`].(int64)),
		Name:   data.params[`name`].(string),
		Value:  data.params[`value`].(string),
	}
	hash, err := sendEmbeddedTx(header.Type, header.UserID, toSerialize)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	data.result = hash
	return nil
}
