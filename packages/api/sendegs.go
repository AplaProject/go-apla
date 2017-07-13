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
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"
)

func preSendEGS(w http.ResponseWriter, r *http.Request, data *apiData) error {

	timeNow := time.Now().Unix()
	forsign := fmt.Sprintf(`%d,%d,%d,`, utils.TypeInt(`DLTTransfer`), timeNow, data.sess.Get(`wallet`).(int64))
	forsign += fmt.Sprintf(`%s,%s,%s,%s`, data.params[`recipient`].(string), data.params[`amount`].(string),
		data.params[`commission`].(string), data.params[`comment`].(string))
	data.result = &forSign{Time: converter.Int64ToStr(timeNow), ForSign: forsign}
	return nil
}

func sendEGS(w http.ResponseWriter, r *http.Request, data *apiData) error {
	header, err := getHeader(`DLTTransfer`, data)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}
	header.StateID = 0

	var toSerialize interface{}

	vcomment := data.params[`comment`].(string)
	if len(vcomment) == 0 {
		vcomment = "null"
	}
	toSerialize = tx.DLTTransfer{
		Header:        header,
		WalletAddress: data.params[`recipient`].(string),
		Amount:        data.params[`amount`].(string),
		Commission:    data.params[`commission`].(string),
		Comment:       vcomment,
	}
	hash, err := sendEmbeddedTx(header.Type, header.UserID, toSerialize)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}
	data.result = hash
	return nil
}
