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

func preSendEGS(w http.ResponseWriter, r *http.Request, data *apiData) error {

	timeNow := time.Now().Unix()
	forsign := fmt.Sprintf(`%d,%d,%d,`, utils.TypeInt(`DLTTransfer`), timeNow, data.sess.Get(`wallet`).(int64))
	forsign += fmt.Sprintf(`%s,%s,%s,%s`, data.params[`recipient`].(string), data.params[`amount`].(string),
		data.params[`commission`].(string), data.params[`comment`].(string))
	data.result = &forSign{Time: converter.Int64ToStr(timeNow), ForSign: forsign}
	return nil
}

func sendEGS(w http.ResponseWriter, r *http.Request, data *apiData) error {
	publicKey := data.params[`pubkey`].([]byte)
	lenpub := len(publicKey)
	if lenpub > 64 {
		publicKey = publicKey[lenpub-64:]
	} else if lenpub == 0 {
		publicKey = []byte("null")
	}

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
	stateID := 0

	var (
		idata []byte
	)
	walletAddress := []byte(data.params[`recipient`].(string))
	amount := []byte(data.params[`amount`].(string))
	commission := []byte(data.params[`commission`].(string))
	vcomment := data.params[`comment`].(string)
	if len(vcomment) == 0 {
		vcomment = "null"
	}

	comment := []byte(vcomment)
	idata = converter.DecToBin(5, 1)
	idata = append(idata, converter.DecToBin(txTime, 4)...)
	idata = append(idata, converter.EncodeLengthPlusData(userID)...)
	idata = append(idata, converter.EncodeLengthPlusData(stateID)...)
	idata = append(idata, converter.EncodeLengthPlusData(walletAddress)...)
	idata = append(idata, converter.EncodeLengthPlusData(amount)...)
	idata = append(idata, converter.EncodeLengthPlusData(commission)...)
	idata = append(idata, converter.EncodeLengthPlusData(comment)...)
	idata = append(idata, converter.EncodeLengthPlusData(publicKey)...)
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
