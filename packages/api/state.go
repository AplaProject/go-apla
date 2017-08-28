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

type stateParamResult struct {
	Name       string `json:"name"`
	Value      string `json:"value"`
	Conditions string `json:"conditions"`
}

type stateParamListResult struct {
	Count string             `json:"count"`
	List  []stateParamResult `json:"list"`
}

type stateItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Logo   string `json:"logo"`
	Coords string `json:"coords"`
}

type stateListResult struct {
	Count string      `json:"count"`
	List  []stateItem `json:"list"`
}

func getStateParams(w http.ResponseWriter, r *http.Request, data *apiData) error {

	dataPar, err := model.GetOneRow(`SELECT * FROM "`+getPrefix(data)+`_state_parameters" WHERE name = ?`,
		data.params[`name`].(string)).String()
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	data.result = &stateParamResult{Name: dataPar["name"], Value: dataPar["value"], Conditions: dataPar["conditions"]}
	return nil
}

func txPreNewState(w http.ResponseWriter, r *http.Request, data *apiData) error {
	v := tx.NewState{
		Header:       getSignHeader(`NewState`, data),
		StateName:    data.params[`name`].(string),
		CurrencyName: data.params[`currency`].(string),
	}
	data.result = &forSign{Time: converter.Int64ToStr(v.Time), ForSign: v.ForSign()}
	return nil
}

func txNewState(w http.ResponseWriter, r *http.Request, data *apiData) error {
	header, err := getHeader(`NewState`, data)
	header.StateID = 0
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusBadRequest)
	}

	toSerialize := tx.NewState{
		Header:       header,
		StateName:    data.params[`name`].(string),
		CurrencyName: data.params[`currency`].(string),
	}
	hash, err := sendEmbeddedTx(header.Type, header.UserID, toSerialize)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	data.result = hash
	return nil
}

func txPreNewStateParams(w http.ResponseWriter, r *http.Request, data *apiData) error {
	v := tx.NewStateParameters{
		Header:     getSignHeader(`NewStateParameters`, data),
		Name:       data.params[`name`].(string),
		Value:      data.params[`value`].(string),
		Conditions: data.params[`conditions`].(string),
	}
	data.result = &forSign{Time: converter.Int64ToStr(v.Time), ForSign: v.ForSign()}
	return nil
}

func txPreEditStateParams(w http.ResponseWriter, r *http.Request, data *apiData) error {
	v := tx.EditStateParameters{
		Header:     getSignHeader(`EditStateParameters`, data),
		Name:       data.params[`name`].(string),
		Value:      data.params[`value`].(string),
		Conditions: data.params[`conditions`].(string),
	}
	data.result = &forSign{Time: converter.Int64ToStr(v.Time), ForSign: v.ForSign()}
	return nil
}

func txStateParams(w http.ResponseWriter, r *http.Request, data *apiData) error {
	txName := `NewStateParameters`
	if r.Method == `PUT` {
		txName = `EditStateParameters`
	}
	header, err := getHeader(txName, data)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusBadRequest)
	}

	var toSerialize interface{}

	if txName == `EditStateParameters` {
		toSerialize = tx.EditStateParameters{
			Header:     header,
			Name:       data.params[`name`].(string),
			Value:      data.params[`value`].(string),
			Conditions: data.params[`conditions`].(string),
		}
	} else {
		toSerialize = tx.NewStateParameters{
			Header:     header,
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

func stateParamsList(w http.ResponseWriter, r *http.Request, data *apiData) error {

	limit := int(data.params[`limit`].(int64))
	if limit == 0 {
		limit = 25
	} else if limit < 0 {
		limit = -1
	}
	outList := make([]stateParamResult, 0)
	count, err := model.Single(`SELECT count(*) FROM "` + getPrefix(data) + `_state_parameters"`).String()
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}

	list, err := model.GetAll(`SELECT * FROM "`+getPrefix(data)+`_state_parameters" order by name`+
		fmt.Sprintf(` offset %d `, data.params[`offset`].(int64)), limit)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}

	for _, val := range list {
		outList = append(outList, stateParamResult{Name: val[`name`], Value: val[`value`],
			Conditions: val[`conditions`]})
	}
	data.result = &stateParamListResult{Count: count, List: outList}
	return nil
}

func stateList(w http.ResponseWriter, r *http.Request, data *apiData) error {
	limit := int(data.params[`limit`].(int64))
	if limit == 0 {
		limit = 25
	} else if limit < 0 {
		limit = -1
	}
	count, err := model.Single(`SELECT count(*) FROM system_states`).String()
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	idata, err := model.GetList(`SELECT id FROM system_states order by id desc` +
		fmt.Sprintf(` offset %d limit %d`, data.params[`offset`].(int64), limit)).String()
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	outList := make([]stateItem, 0)
	for _, id := range idata {
		if !model.IsNodeState(converter.StrToInt64(id), r.Host) {
			continue
		}
		list, err := model.GetAll(fmt.Sprintf(`SELECT name, value FROM "%s_state_parameters" WHERE name in ('state_name','state_flag', 'state_coords')`,
			id), -1)
		if err != nil {
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}
		item := stateItem{ID: id}
		for _, val := range list {
			switch val[`name`] {
			case `state_name`:
				item.Name = val[`value`]
			case `state_flag`:
				item.Logo = val[`value`]
			case `state_coords`:
				item.Coords = val[`value`]
			}
		}
		outList = append(outList, item)
	}
	data.result = &stateListResult{Count: count, List: outList}
	return nil
}
