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
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"
)

type pageResult struct {
	Name       string `json:"name"`
	Menu       string `json:"menu"`
	Value      string `json:"value"`
	Conditions string `json:"conditions"`
}

type pageItem struct {
	Name string `json:"name"`
	Menu string `json:"menu"`
}

type pageListResult struct {
	Count string     `json:"count"`
	List  []pageItem `json:"list"`
}

func getPage(w http.ResponseWriter, r *http.Request, data *apiData) error {

	dataPage, err := model.GetOneRow(`SELECT * FROM "`+getPrefix(data)+`_pages" WHERE name = ?`,
		data.params[`name`].(string)).String()
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}
	data.result = &pageResult{Name: dataPage["name"], Menu: dataPage["menu"],
		Value: dataPage["value"], Conditions: dataPage["conditions"]}
	return nil
}

func txPreNewPage(w http.ResponseWriter, r *http.Request, data *apiData) error {
	v := tx.NewPage{
		Header:     getSignHeader(`NewPage`, data),
		Global:     converter.Int64ToStr(data.params[`global`].(int64)),
		Name:       data.params[`name`].(string),
		Value:      data.params[`value`].(string),
		Menu:       data.params[`menu`].(string),
		Conditions: data.params[`conditions`].(string),
	}
	data.result = &forSign{Time: converter.Int64ToStr(v.Time), ForSign: v.ForSign()}
	return nil
}

func txPreEditPage(w http.ResponseWriter, r *http.Request, data *apiData) error {
	v := tx.EditPage{
		Header:     getSignHeader(`EditPage`, data),
		Global:     converter.Int64ToStr(data.params[`global`].(int64)),
		Name:       data.params[`name`].(string),
		Value:      data.params[`value`].(string),
		Menu:       data.params[`menu`].(string),
		Conditions: data.params[`conditions`].(string),
	}
	data.result = &forSign{Time: converter.Int64ToStr(v.Time), ForSign: v.ForSign()}
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

func pageList(w http.ResponseWriter, r *http.Request, data *apiData) error {

	limit := int(data.params[`limit`].(int64))
	if limit == 0 {
		limit = 25
	} else if limit < 0 {
		limit = -1
	}
	outList := make([]pageItem, 0)
	count, err := model.Single(`SELECT count(*) FROM "` + getPrefix(data) + `_pages"`).String()
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}

	list, err := model.GetAll(`SELECT name, menu FROM "`+getPrefix(data)+`_pages" order by name`+
		fmt.Sprintf(` offset %d `, data.params[`offset`].(int64)), limit)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}

	for _, val := range list {
		outList = append(outList, pageItem{Name: val[`name`], Menu: val[`menu`]})
	}
	data.result = &pageListResult{Count: count, List: outList}
	return nil
}

func txPreAppendPage(w http.ResponseWriter, r *http.Request, data *apiData) error {
	v := tx.AppendPage{
		Header: getSignHeader(`AppendPage`, data),
		Global: converter.Int64ToStr(data.params[`global`].(int64)),
		Name:   data.params[`name`].(string),
		Value:  data.params[`value`].(string),
	}
	data.result = &forSign{Time: converter.Int64ToStr(v.Time), ForSign: v.ForSign()}
	return nil
}

func txAppendPage(w http.ResponseWriter, r *http.Request, data *apiData) error {
	header, err := getHeader(`AppendPage`, data)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}

	toSerialize := tx.AppendPage{
		Header: header,
		Global: converter.Int64ToStr(data.params[`global`].(int64)),
		Name:   data.params[`name`].(string),
		Value:  data.params[`value`].(string),
	}
	hash, err := sendEmbeddedTx(header.Type, header.UserID, toSerialize)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}
	data.result = hash

	return nil
}
