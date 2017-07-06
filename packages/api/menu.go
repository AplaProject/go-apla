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
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/sql"
)

type menuResult struct {
	Name       string `json:"name"`
	Value      string `json:"value"`
	Conditions string `json:"conditions"`
}

func getMenu(w http.ResponseWriter, r *http.Request, data *apiData) error {

	dataMenu, err := sql.DB.OneRow(`SELECT * FROM "`+getPrefix(data)+`_menu" WHERE name = ?`,
		data.params[`name`].(string)).String()
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}
	data.result = &menuResult{Name: dataMenu["name"], Value: dataMenu["value"], Conditions: dataMenu["conditions"]}
	return nil
}

func prePostMenu(w http.ResponseWriter, r *http.Request, data *apiData) error {
	timeNow := time.Now().Unix()
	forsign := fmt.Sprintf(`%d,%d,%d,%d,`, utils.TypeInt(`NewMenu`), timeNow, data.sess.Get(`citizen`).(int64),
		data.sess.Get(`state`).(int64))
	forsign += fmt.Sprintf(`%d,%s,%s,%s`, data.params[`global`].(int64), data.params[`name`].(string),
		data.params[`value`].(string), data.params[`conditions`].(string))
	data.result = &forSign{Time: converter.Int64ToStr(timeNow), ForSign: forsign}
	return nil
}

func postMenu(w http.ResponseWriter, r *http.Request, data *apiData) error {

	txTime := converter.StrToInt64(data.params[`time`].(string))
	sign := make([]byte, 0)
	signature := data.params[`signature`].([]byte)
	if len(signature) > 0 {
		sign = append(sign, converter.EncodeLengthPlusData(signature)...)
	}
	if len(sign) == 0 {
		return errorAPI(w, "signature is empty", http.StatusConflict)
	}
	binSignatures := converter.EncodeLengthPlusData(sign)

	userID := data.sess.Get(`wallet`).(int64)
	stateID := data.sess.Get(`state`).(int64)

	var (
		idata []byte
	)
	global := []byte(converter.Int64ToStr(data.params[`global`].(int64)))
	name := []byte(data.params[`name`].(string))
	value := []byte(data.params[`value`].(string))
	conditions := []byte(data.params[`conditions`].(string))

	idata = converter.DecToBin(21, 1)

	idata = append(idata, converter.DecToBin(txTime, 4)...)
	idata = append(idata, converter.EncodeLengthPlusData(userID)...)
	idata = append(idata, converter.EncodeLengthPlusData(stateID)...)
	idata = append(idata, converter.EncodeLengthPlusData(global)...)
	idata = append(idata, converter.EncodeLengthPlusData(name)...)
	idata = append(idata, converter.EncodeLengthPlusData(value)...)
	idata = append(idata, converter.EncodeLengthPlusData(conditions)...)
	idata = append(idata, binSignatures...)

	hash, err := crypto.Hash(idata)

	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}

	hash = converter.BinToHex(hash)
	err = sql.DB.ExecSQL(`INSERT INTO transactions_status (
				hash,
				time,
				type,
				wallet_id,
				citizen_id
			)
			VALUES (
				[hex],
				?,
				?,
				?,
				?
			)`, hash, time.Now().Unix(), 5, userID, userID)

	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}

	err = sql.DB.ExecSQL("INSERT INTO queue_tx (hash, data) VALUES ([hex], [hex])", hash, converter.BinToHex(idata))
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}
	data.result = &hashTx{Hash: string(hash)}

	return nil
}

func prePutMenu(w http.ResponseWriter, r *http.Request, data *apiData) error {
	timeNow := time.Now().Unix()
	forsign := fmt.Sprintf(`%d,%d,%d,%d,`, utils.TypeInt(`EditMenu`), timeNow, data.sess.Get(`citizen`).(int64),
		data.sess.Get(`state`).(int64))
	forsign += fmt.Sprintf(`%d,%s,%s,%s`, data.params[`global`].(int64), data.params[`name`].(string),
		data.params[`value`].(string), data.params[`conditions`].(string))
	data.result = &forSign{Time: converter.Int64ToStr(timeNow), ForSign: forsign}
	return nil
}

func putMenu(w http.ResponseWriter, r *http.Request, data *apiData) error {

	txTime := converter.StrToInt64(data.params[`time`].(string))
	sign := make([]byte, 0)
	signature := data.params[`signature`].([]byte)
	if len(signature) > 0 {
		sign = append(sign, converter.EncodeLengthPlusData(signature)...)
	}
	if len(sign) == 0 {
		return errorAPI(w, "signature is empty", http.StatusConflict)
	}
	binSignatures := converter.EncodeLengthPlusData(sign)

	userID := data.sess.Get(`wallet`).(int64)
	stateID := data.sess.Get(`state`).(int64)

	var (
		idata []byte
	)
	global := []byte(converter.Int64ToStr(data.params[`global`].(int64)))
	name := []byte(data.params[`name`].(string))
	value := []byte(data.params[`value`].(string))
	conditions := []byte(data.params[`conditions`].(string))

	idata = converter.DecToBin(13, 1)

	idata = append(idata, converter.DecToBin(txTime, 4)...)
	idata = append(idata, converter.EncodeLengthPlusData(userID)...)
	idata = append(idata, converter.EncodeLengthPlusData(stateID)...)
	idata = append(idata, converter.EncodeLengthPlusData(global)...)
	idata = append(idata, converter.EncodeLengthPlusData(name)...)
	idata = append(idata, converter.EncodeLengthPlusData(value)...)
	idata = append(idata, converter.EncodeLengthPlusData(conditions)...)
	idata = append(idata, binSignatures...)

	hash, err := crypto.Hash(idata)

	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}

	hash = converter.BinToHex(hash)
	err = sql.DB.ExecSQL(`INSERT INTO transactions_status (
				hash,
				time,
				type,
				wallet_id,
				citizen_id
			)
			VALUES (
				[hex],
				?,
				?,
				?,
				?
			)`, hash, time.Now().Unix(), 5, userID, userID)

	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}

	err = sql.DB.ExecSQL("INSERT INTO queue_tx (hash, data) VALUES ([hex], [hex])", hash, converter.BinToHex(idata))
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}
	data.result = &hashTx{Hash: string(hash)}

	return nil
}
