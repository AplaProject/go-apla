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

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
)

const aEncryptKey = `ajax_encrypt_key`

// EncryptKey is a structure for the answer of ajax_encrypt_key ajax request
type EncryptKey struct {
	Encrypted string `json:"encrypted"` //hex
	Public    string `json:"public"`    //hex
	WalletID  int64  `json:"wallet_id"`
	Address   string `json:"address"`
	Error     string `json:"error"`
}

func init() {
	newPage(aEncryptKey, `json`)
}

// EncryptNewKey creates a shared key, generates and crypts a new private key
func EncryptNewKey(walletID string) (result EncryptKey) {
	var (
		err error
		id  int64
	)

	if len(walletID) == 0 {
		result.Error = `unknown wallet id`
		return result
	}
	id = converter.StringToAddress(walletID)
	wallet := &model.DltWallet{}
	err = wallet.GetWallet(id)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if len(wallet.PublicKey) == 0 {
		result.Error = `unknown wallet id`
		return result
	}
	var private string

	for result.WalletID == 0 {
		private, result.Public, _ = crypto.GenHexKeys()

		pub, _ := hex.DecodeString(result.Public)

		wallet.WalletID = crypto.Address(pub)
		exist, err := wallet.IsExists()
		if err != nil {
			result.Error = err.Error()
			return result
		}
		if exist == false {
			result.WalletID = wallet.WalletID
		}
	}
	priv, _ := hex.DecodeString(private)
	encrypted, err := crypto.SharedEncrypt([]byte(wallet.PublicKey), priv)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.Encrypted = hex.EncodeToString(encrypted)
	result.Address = converter.AddressToString(result.WalletID)

	return
}

// AjaxEncryptKey is a controller of ajax_encrypt_key request
func (c *Controller) AjaxEncryptKey() interface{} {
	return EncryptNewKey(c.r.FormValue("wallet_id"))
}
