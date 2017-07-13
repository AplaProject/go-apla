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

type contractResult struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Active     string `json:"active"`
	Wallet     string `json:"wallet"`
	Value      string `json:"value"`
	Conditions string `json:"conditions"`
}

type contractItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Active string `json:"active"`
	Wallet string `json:"wallet"`
}

type contractListResult struct {
	Count string         `json:"count"`
	List  []contractItem `json:"list"`
}

func checkID(data *apiData) (id string, err error) {
	id = data.params[`id`].(string)
	if id[0] > '9' {
		id, err = sql.DB.Single(`SELECT id FROM "`+getPrefix(data)+`_smart_contracts" WHERE name = ?`, id).String()
		if err == nil && len(id) == 0 {
			err = fmt.Errorf(`incorrect id %s of the contract`, data.params[`id`].(string))
		}
	}
	return
}

func getContract(w http.ResponseWriter, r *http.Request, data *apiData) error {
	id, err := checkID(data)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusBadRequest)
	}
	dataContract, err := sql.DB.OneRow(`SELECT * FROM "`+getPrefix(data)+`_smart_contracts" WHERE id = ?`, id).String()
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}
	data.result = &contractResult{ID: dataContract["id"], Name: dataContract["name"], Active: dataContract["active"],
		Wallet: dataContract["wallet"], Value: dataContract["value"], Conditions: dataContract["conditions"]}
	return nil
}

func txPreContract(w http.ResponseWriter, r *http.Request, data *apiData) error {
	var name, par string

	if r.Method == `PUT` {
		name = `EditContract`
		par = `id`
	} else {
		name = `NewContract`
		par = `name`
		if data.params[`wallet`].(int64) > 0 {
			data.params[par] = fmt.Sprintf(`%s#%d`, data.params[par].(string), data.params[`wallet`].(int64))
		}
	}
	forsign := fmt.Sprintf(`%d,%v,%s,%s`, data.params[`global`].(int64), data.params[par].(string),
		data.params[`value`].(string), data.params[`conditions`].(string))
	data.result = getForSign(name, data, forsign)
	return nil
}

func txContract(w http.ResponseWriter, r *http.Request, data *apiData) error {
	var txName string

	if r.Method == `PUT` {
		txName = `EditContract`
		id, err := checkID(data)
		if err != nil {
			return errorAPI(w, err.Error(), http.StatusBadRequest)
		}
		data.params[`name`] = id
	} else {
		txName = `NewContract`
		if data.params[`wallet`].(int64) > 0 {
			data.params[`name`] = fmt.Sprintf(`%s#%d`, data.params[`name`].(string), data.params[`wallet`].(int64))
		}
	}
	header, err := getHeader(txName, data)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}

	var toSerialize interface{}

	if txName == `EditContract` {
		toSerialize = tx.EditContract{
			Header:     header,
			Global:     converter.Int64ToStr(data.params[`global`].(int64)),
			Id:         data.params[`name`].(string),
			Value:      data.params[`value`].(string),
			Conditions: data.params[`conditions`].(string),
		}
	} else {
		toSerialize = tx.NewContract{
			Header:     header,
			Global:     converter.Int64ToStr(data.params[`global`].(int64)),
			Name:       data.params[`name`].(string),
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

func txPreActivateContract(w http.ResponseWriter, r *http.Request, data *apiData) error {
	var global int64
	id, err := checkID(data)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusBadRequest)
	}
	if _, ok := data.params[`global`]; ok {
		global = 1
	}
	data.result = getForSign(`ActivateContract`, data, fmt.Sprintf(`%d,%s`, global, id))
	return nil
}

func txActivateContract(w http.ResponseWriter, r *http.Request, data *apiData) error {

	txName := `ActivateContract`
	id, err := checkID(data)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusBadRequest)
	}
	header, err := getHeader(txName, data)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}

	var (
		global      string
		toSerialize interface{}
	)
	if _, ok := data.params[`global`]; ok {
		global = `1`
	} else {
		global = `0`
	}
	toSerialize = tx.ActivateContract{
		Header: header,
		Global: global,
		Id:     id,
	}
	hash, err := sendEmbeddedTx(header.Type, header.UserID, toSerialize)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}
	data.result = hash
	return nil
}

func contractList(w http.ResponseWriter, r *http.Request, data *apiData) error {

	limit := -1
	if val, ok := data.params[`limit`]; ok {
		limit = converter.StrToInt(val.(string))
	}
	outList := make([]contractItem, 0)
	count, err := sql.DB.Single(`SELECT count(*) FROM "` + getPrefix(data) + `_smart_contracts"`).String()
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}

	if limit != 0 {
		list, err := sql.DB.GetAll(`SELECT * FROM "`+getPrefix(data)+`_smart_contracts" order by id`, limit)
		if err != nil {
			return errorAPI(w, err.Error(), http.StatusConflict)
		}

		for _, val := range list {
			var wallet, active string
			if val[`wallet_id`] != `NULL` {
				wallet = converter.AddressToString(converter.StrToInt64(val[`wallet_id`]))
			}
			if val[`active`] != `NULL` {
				active = `1`
			}
			outList = append(outList, contractItem{ID: val[`id`], Name: val[`name`], Wallet: wallet, Active: active})
		}
	}
	data.result = &contractListResult{Count: count, List: outList}
	return nil
}
