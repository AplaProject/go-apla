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

package controllers

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"io/ioutil"
)

const ANewState = `ajax_new_state`

type NewState struct {
	Error string `json:"error"`
}

func init() {
	newPage(ANewState, `json`)
}

func (c *Controller) AjaxNewState() interface{} {
	var (
		result    NewKey
		err       error
		spriv     string
		priv, pub []byte
		wallet    int64
	)
	id := utils.StrToInt64(c.r.FormValue("testnet"))
	if current, err := c.OneRow(`select wallet, private from testnet_emails where id=?`, id).String(); err != nil {
		result.Error = err.Error()
	} else if len(current) == 0 {
		result.Error = `unknown id`
	} else if utils.StrToInt64(current[`wallet`]) > 0 || len(current[`private`]) > 0 {
		result.Error = `duplicate of request`
	}
	if len(result.Error) > 0 {
		return result
	}
	exist := int64(1)
	for exist != 0 {
		spriv, _ = lib.GenKeys()
		priv, _ = hex.DecodeString(spriv)
		pub = lib.PrivateToPublic(priv)
		wallet = int64(lib.Address(pub))

		exist, err = c.Single(`select wallet_id from dlt_wallets where wallet_id=?`, wallet).Int64()
		if err != nil {
			result.Error = err.Error()
			return result
		}
	}
	adminKey, err := ioutil.ReadFile(*utils.Dir + `/TestnetKey`)
	if err != nil || len(strings.TrimSpace(string(adminKey))) != 64 {
		result.Error = `TestnetKey is absent`
		return result
	}
	err = c.ExecSql(`update testnet_emails set wallet=?, private=? where id=?`, wallet, spriv, id)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	adminPriv, err := hex.DecodeString(string(adminKey))
	if err != nil {
		result.Error = err.Error()
		return result
	}
	adminWallet := int64(lib.Address(lib.PrivateToPublic(adminPriv)))
	walletUser := strings.Replace(lib.AddressToString(uint64(wallet)), `-`, ``, -1)

	txType := utils.TypeInt(`DLTTransfer`)
	txTime := time.Now().Unix()
	forSign := fmt.Sprintf("%d,%d,%d,%s,%s,%s,%s", txType, txTime, adminWallet,
		walletUser, `1e+21`, `1000000000000000`, `testnet`)
	signature, err := lib.SignECDSA(string(adminKey), forSign)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	/*	fmt.Println(`FORIN`, forSign)
		fmt.Println(`IN`, hex.EncodeToString(signature))*/
	sign := make([]byte, 0)
	sign = append(sign, utils.EncodeLengthPlusData(signature)...)
	binsign := utils.EncodeLengthPlusData(sign)

	data := make([]byte, 0)
	data = utils.DecToBin(txType, 1)
	data = append(data, utils.DecToBin(txTime, 4)...)
	data = append(data, utils.EncodeLengthPlusData(adminWallet)...)
	data = append(data, utils.EncodeLengthPlusData(0)...)
	data = append(data, utils.EncodeLengthPlusData([]byte(walletUser))...)
	data = append(data, utils.EncodeLengthPlusData([]byte(`1e+21`))...)
	data = append(data, utils.EncodeLengthPlusData([]byte(`1000000000000000`))...)
	data = append(data, utils.EncodeLengthPlusData([]byte(`testnet`))...)
	data = append(data, utils.EncodeLengthPlusData([]byte(``))...)
	data = append(data, binsign...)

	md5 := utils.Md5(data)
	err = c.ExecSql(`INSERT INTO transactions_status (
			hash, time,	type, wallet_id, citizen_id	) VALUES (
			[hex], ?, ?, ?, ? )`, md5, time.Now().Unix(), txType, adminWallet, adminWallet)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	err = c.ExecSql("INSERT INTO queue_tx (hash, data) VALUES ([hex], [hex])", md5, hex.EncodeToString(data))
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.Error = `success`

	return result
}
