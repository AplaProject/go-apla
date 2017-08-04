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

	"io/ioutil"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const aNewState = `ajax_new_state`

// NewState is a structure for the answer of ajax_new_state ajax request
type NewStateAjax struct {
	Error string `json:"error"`
}

func init() {
	newPage(aNewState, `json`)
}

// AjaxNewState is a controller of ajax_new_state request
func (c *Controller) AjaxNewState() interface{} {
	var (
		result    NewStateAjax
		err       error
		spriv     string
		current   map[string]string
		priv, pub []byte
		wallet    int64
	)
	id := converter.StrToInt64(c.r.FormValue("testnet"))
	if current, err = c.OneRow(`select country,currency,wallet, private from testnet_emails where id=?`, id).String(); err != nil {
		result.Error = err.Error()
	} else if len(current) == 0 {
		result.Error = `unknown id`
	} else if converter.StrToInt64(current[`wallet`]) > 0 || len(current[`private`]) > 0 {
		result.Error = `duplicate of request`
	}
	if len(result.Error) > 0 {
		return result
	}
	exist := int64(1)
	for exist != 0 {
		spriv, _, _ = crypto.GenHexKeys()
		priv, _ = hex.DecodeString(spriv)
		pub, err = crypto.PrivateToPublic(priv)
		if err != nil {
			log.Fatal(err)
		}
		wallet = crypto.Address(pub)

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
	err = c.ExecSQL(`update testnet_emails set wallet=?, private=? where id=?`, wallet, spriv, id)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	adminPriv, err := hex.DecodeString(string(adminKey))
	if err != nil {
		result.Error = err.Error()
		return result
	}
	publicKey, err := crypto.PrivateToPublic(adminPriv)
	if err != nil {
		log.Fatal(err)
	}
	adminWallet := crypto.Address(publicKey)
	walletUser := strings.Replace(converter.AddressToString(wallet), `-`, ``, -1)

	txType := utils.TypeInt(`DLTTransfer`)
	txTime := time.Now().Unix()
	forSign := fmt.Sprintf("%d,%d,%d,%s,%s,%s,%s", txType, txTime, adminWallet,
		walletUser, `2e+21`, `1000000000000000`, `testnet`)
	signature, err := crypto.Sign(string(adminKey), forSign)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	/*	fmt.Println(`FORIN`, forSign)
		fmt.Println(`IN`, hex.EncodeToString(signature))*/
	sign := make([]byte, 0)
	sign = append(sign, converter.EncodeLengthPlusData(signature)...)
	binsign := converter.EncodeLengthPlusData(sign)

	data := make([]byte, 0)
	data = converter.DecToBin(txType, 1)
	data = append(data, converter.DecToBin(txTime, 4)...)
	data = append(data, converter.EncodeLengthPlusData(adminWallet)...)
	data = append(data, converter.EncodeLengthPlusData(0)...)
	data = append(data, converter.EncodeLengthPlusData([]byte(walletUser))...)
	data = append(data, converter.EncodeLengthPlusData([]byte(`2e+21`))...)
	data = append(data, converter.EncodeLengthPlusData([]byte(`1000000000000000`))...)
	data = append(data, converter.EncodeLengthPlusData([]byte(`testnet`))...)
	data = append(data, converter.EncodeLengthPlusData([]byte(``))...)
	data = append(data, binsign...)

	_, err = c.SendTx(txType, adminWallet, data)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	time.Sleep(2500 * time.Millisecond)
	txType = utils.TypeInt(`NewState`)
	txTime = time.Now().Unix()
	forSign = fmt.Sprintf("%d,%d,%d,%s,%s", txType, txTime, wallet, current[`country`], current[`currency`])
	signature, err = crypto.Sign(spriv, forSign)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	sign = converter.EncodeLengthPlusData(signature)
	binsign = converter.EncodeLengthPlusData(sign)
	data = data[:0]
	data = converter.DecToBin(txType, 1)
	data = append(data, converter.DecToBin(txTime, 4)...)
	data = append(data, converter.EncodeLengthPlusData(wallet)...)
	data = append(data, converter.EncodeLengthPlusData(0)...)
	data = append(data, converter.EncodeLengthPlusData([]byte(current[`country`]))...)
	data = append(data, converter.EncodeLengthPlusData([]byte(current[`currency`]))...)
	data = append(data, converter.EncodeLengthPlusData(hex.EncodeToString(pub))...)
	data = append(data, binsign...)
	/*	pubkey := make([][]byte, 0)
		pubkey = append(pubkey, pub)
		CheckSignResult, err := utils.CheckSign(pubkey, forSign, sign, false)
		fmt.Println(`CHECK`, CheckSignResult, err)*/

	_, err = c.SendTx(txType, wallet, data)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.Error = `success`

	return result
}
