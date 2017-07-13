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

type pageResult struct {
	Name       string `json:"name"`
	Menu       string `json:"menu"`
	Value      string `json:"value"`
	Conditions string `json:"conditions"`
}

func getPage(w http.ResponseWriter, r *http.Request, data *apiData) error {

	dataPage, err := sql.DB.OneRow(`SELECT * FROM "`+getPrefix(data)+`_pages" WHERE name = ?`,
		data.params[`name`].(string)).String()
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}
	data.result = &pageResult{Name: dataPage["name"], Menu: dataPage["menu"],
		Value: dataPage["value"], Conditions: dataPage["conditions"]}
	return nil
}

func txPrePage(w http.ResponseWriter, r *http.Request, data *apiData) error {
	name := `NewPage`
	if r.Method == `PUT` {
		name = `EditPage`
	}
	forsign := fmt.Sprintf(`%d,%s,%s,%s,%s`, data.params[`global`].(int64), data.params[`name`].(string),
		data.params[`value`].(string), data.params[`menu`].(string), data.params[`conditions`].(string))
	data.result = getForSign(name, data, forsign)
	return nil
}

func txPage(w http.ResponseWriter, r *http.Request, data *apiData) error {
	txName := `NewPage`
	if r.Method == `PUT` {
		txName = `EditPage`
	}
	header, err := getHeader(txName, data)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}

	var toSerialize interface{}

	if txName == `EditPage` {
		toSerialize = tx.EditPage{
			Header:     header,
			Global:     converter.Int64ToStr(data.params[`global`].(int64)),
			Name:       data.params[`name`].(string),
			Menu:       data.params[`menu`].(string),
			Value:      data.params[`value`].(string),
			Conditions: data.params[`conditions`].(string),
		}
	} else {
		toSerialize = tx.NewPage{
			Header:     header,
			Global:     converter.Int64ToStr(data.params[`global`].(int64)),
			Name:       data.params[`name`].(string),
			Menu:       data.params[`menu`].(string),
			Value:      data.params[`value`].(string),
			Conditions: data.params[`conditions`].(string),
		}
	}
	hash, err := sendEmbeddedTx(header.Type, header.UserID, toSerialize)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}
	data.result = hash

	return nil
}
