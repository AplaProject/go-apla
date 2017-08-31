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
	"strconv"
	"strings"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"

	"io/ioutil"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
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
		priv, pub []byte
		wallet    int64
	)
	id, err := strconv.ParseInt(c.r.FormValue("testnet"), 10, 64)
	if err != nil {
		logger.LogInfo(consts.StrToIntError, c.r.FormValue("testnet"))
	}
	testnetEmail := &model.TestnetEmail{ID: id}
	if err = testnetEmail.Get(id); err != nil {
		result.Error = err.Error()
	} else if testnetEmail.Wallet > 0 || len(testnetEmail.Private) > 0 {
		result.Error = `duplicate of request`
	}
	if len(result.Error) > 0 {
		return result
	}
	exist := true
	for exist != false {
		spriv, _, _ = crypto.GenHexKeys()
		priv, _ = hex.DecodeString(spriv)
		pub, err = crypto.PrivateToPublic(priv)
		if err != nil {
			log.Fatal(err)
		}
		wallet = crypto.Address(pub)

		dltWallet := &model.DltWallet{WalletID: wallet}
		exist, err = dltWallet.IsExists()
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
	testnetEmail.Wallet = wallet
	testnetEmail.Private = priv

	err = testnetEmail.Save()
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

	_, err = model.SendTx(txType, adminWallet, data)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	time.Sleep(2500 * time.Millisecond)
	txType = utils.TypeInt(`NewState`)
	txTime = time.Now().Unix()
	forSign = fmt.Sprintf("%d,%d,%d,%s,%s", txType, txTime, wallet, testnetEmail.Country, testnetEmail.Currency)
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
	data = append(data, converter.EncodeLengthPlusData([]byte(testnetEmail.Country))...)
	data = append(data, converter.EncodeLengthPlusData([]byte(testnetEmail.Currency))...)
	data = append(data, converter.EncodeLengthPlusData(hex.EncodeToString(pub))...)
	data = append(data, binsign...)

	_, err = model.SendTx(txType, wallet, data)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.Error = `success`

	return result
}
