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
	"strconv"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
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
		id, err = model.Single(`SELECT id FROM "`+getPrefix(data)+`_smart_contracts" WHERE name = ?`, id).String()
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
	dataContract, err := model.GetOneRow(`SELECT * FROM "`+getPrefix(data)+`_smart_contracts" WHERE id = ?`, id).String()
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	data.result = &contractResult{ID: dataContract["id"], Name: dataContract["name"], Active: dataContract["active"],
		Wallet: dataContract["wallet"], Value: dataContract["value"], Conditions: dataContract["conditions"]}
	return nil
}

func txPreNewContract(w http.ResponseWriter, r *http.Request, data *apiData) error {
	v := tx.NewContract{
		Header:     getSignHeader(`NewContract`, data),
		Global:     converter.Int64ToStr(data.params[`global`].(int64)),
		Name:       data.params[`name`].(string),
		Value:      data.params[`value`].(string),
		Conditions: data.params[`conditions`].(string),
		Wallet:     data.params[`wallet`].(string),
	}
	data.result = &forSign{Time: converter.Int64ToStr(v.Time), ForSign: v.ForSign()}
	return nil
}

func txPreEditContract(w http.ResponseWriter, r *http.Request, data *apiData) error {
	v := tx.EditContract{
		Header:     getSignHeader(`EditContract`, data),
		Global:     converter.Int64ToStr(data.params[`global`].(int64)),
		Id:         data.params[`id`].(string),
		Value:      data.params[`value`].(string),
		Conditions: data.params[`conditions`].(string),
	}
	data.result = &forSign{Time: converter.Int64ToStr(v.Time), ForSign: v.ForSign()}
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
	}
	header, err := getHeader(txName, data)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusBadRequest)
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
			Wallet:     data.params[`wallet`].(string),
		}
	}
	hash, err := sendEmbeddedTx(header.Type, header.UserID, toSerialize)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	data.result = hash
	return nil
}

func txPreActivateContract(w http.ResponseWriter, r *http.Request, data *apiData) error {
	id, err := checkID(data)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusBadRequest)
	}
	v := tx.ActivateContract{
		Header: getSignHeader(`ActivateContract`, data),
		Global: converter.Int64ToStr(data.params[`global`].(int64)),
		Id:     id,
	}
	data.result = &forSign{Time: converter.Int64ToStr(v.Time), ForSign: v.ForSign()}
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
		return errorAPI(w, err.Error(), http.StatusBadRequest)
	}

	var toSerialize interface{}

	toSerialize = tx.ActivateContract{
		Header: header,
		Global: converter.Int64ToStr(data.params[`global`].(int64)),
		Id:     id,
	}
	hash, err := sendEmbeddedTx(header.Type, header.UserID, toSerialize)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	data.result = hash
	return nil
}

func contractList(w http.ResponseWriter, r *http.Request, data *apiData) error {

	limit := int(data.params[`limit`].(int64))
	if limit == 0 {
		limit = 25
	} else if limit < 0 {
		limit = -1
	}
	outList := make([]contractItem, 0)
	count, err := model.Single(`SELECT count(*) FROM "` + getPrefix(data) + `_smart_contracts"`).String()
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}

	list, err := model.GetAll(`SELECT * FROM "`+getPrefix(data)+`_smart_contracts" order by id`+
		fmt.Sprintf(` offset %d `, data.params[`offset`].(int64)), limit)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}

	for _, val := range list {
		var wallet, active string
		if val[`wallet_id`] != `NULL` {
			walletID, err := strconv.ParseInt(val[`wallet_id`], 10, 64)
			if err != nil {
				logger.LogInfo(consts.StrtoInt64Error, val[`wallet_id`])
			}
			wallet = converter.AddressToString(walletID)

		}
		if val[`active`] != `NULL` {
			active = `1`
		}
		outList = append(outList, contractItem{ID: val[`id`], Name: val[`name`], Wallet: wallet, Active: active})
	}
	data.result = &contractListResult{Count: count, List: outList}
	return nil
}
