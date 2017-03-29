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
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const AEncryptKey = `ajax_encrypt_key`

type EncryptKey struct {
	Encrypted string `json:"encrypted"` //hex
	Public    string `json:"public"`    //hex
	WalletId  int64  `json:"wallet_id"`
	Address   string `json:"address"`
	Error     string `json:"error"`
}

func init() {
	newPage(AEncryptKey, `json`)
}

func (c *Controller) AjaxEncryptKey() interface{} {
	var (
		err    error
		id     int64
		result EncryptKey
	)

	wallet_id := c.r.FormValue("wallet_id")
	if len(wallet_id) == 0 {
		result.Error = `unknown wallet id`
		return result
	}
	if wallet_id[0] == '-' {
		id = utils.StrToInt64(wallet_id)
	} else {
		id = lib.StringToAddress(wallet_id)
	}
	pubKey, err := c.Single(`select public_key_0 from dlt_wallets where wallet_id=?`, id).String()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if len(pubKey) == 0 {
		result.Error = `unknown wallet id`
		return result
	}
	var private string

	for result.WalletId == 0 {
		private, result.Public = lib.GenKeys()

		pub, _ := hex.DecodeString(result.Public)
		idnew := int64(lib.Address(pub))

		exist, err := c.Single(`select wallet_id from dlt_wallets where wallet_id=?`, idnew).Int64()
		if err != nil {
			result.Error = err.Error()
			return result
		}
		if exist == 0 {
			result.WalletId = idnew
		}
	}
	fmt.Println(`AJAX Private`, private)
	priv, _ := hex.DecodeString(private)
	encrypted, err := utils.EncryptShared([]byte(pubKey), priv)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.Encrypted = hex.EncodeToString(encrypted)

	return result
}
