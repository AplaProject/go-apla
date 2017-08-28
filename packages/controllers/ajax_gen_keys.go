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
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"
	"gopkg.in/vmihailenco/msgpack.v2"
)

const aGenKeys = `ajax_gen_keys`

// GenKeys is a structure for the answer of ajax_gen_keys ajax request
type GenKeys struct {
	Generated int64  `json:"generated"`
	Used      int64  `json:"used"`
	Available int64  `json:"available"`
	Error     string `json:"error"`
}

func init() {
	newPage(aGenKeys, `json`)
}

// AjaxGenKeys is a controller of ajax_gen_keys request
func (c *Controller) AjaxGenKeys() interface{} {
	var result GenKeys
	var err error

	count, err := strconv.ParseInt(c.r.FormValue("count"), 10, 64)
	if err != nil {
		logger.LogInfo(consts.StrtoInt64Error, c.r.FormValue("count"))
	}
	if count < 1 || count > 50 {
		result.Error = `Count must be from 1 to 50`
		return result
	}
	stateParameter := &model.StateParameter{}
	stateParameter.SetTablePrefix(converter.Int64ToStr(c.SessStateID))
	err = stateParameter.GetByName("gov_account")
	if err != nil {
		result.Error = err.Error()
		return result
	}
	govAccount, err := strconv.ParseInt(stateParameter.Value, 10, 64)
	if err != nil {
		logger.LogInfo(consts.StrtoInt64Error, stateParameter.Value)
	}
	if c.SessCitizenID != govAccount || len(stateParameter.Value) == 0 {
		result.Error = `Access denied`
		return result
	}

	testnetKey := &model.TestnetKey{}
	err = testnetKey.GetByWallet(c.SessCitizenID)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if len(testnetKey.Private) == 0 {
		result.Error = `unknown private key`
		return result
	}

	contract := smart.GetContract(`GenCitizen`, uint32(c.SessStateID))
	if contract == nil {
		result.Error = `GenCitizen contract has not been found`
		return result
	}

	for i := int64(0); i < count; i++ {
		var priv []byte
		spriv, _, _ := crypto.GenHexKeys()
		priv, _ = hex.DecodeString(spriv)

		pub, err := crypto.PrivateToPublic(priv)
		if err != nil {
			log.Fatal(err)
		}
		idnew := int64(crypto.Address(pub))

		wallet := &model.DltWallet{}
		wallet.WalletID = idnew
		exist, err := wallet.IsExists()
		if err != nil {
			result.Error = err.Error()
			return result
		}
		if exist != false {
			i--
			continue
		}

		ctime := time.Now().Unix()
		info := (*contract).Block.Info.(*script.ContractInfo)
		toSerialize := tx.SmartContract{
			Header: tx.Header{Type: int(info.ID), Time: ctime,
				UserID: c.SessCitizenID, StateID: c.SessStateID}}
		pubhex := hex.EncodeToString(pub)
		forsign := toSerialize.ForSign() + fmt.Sprintf("%v,%v", ``, pubhex)
		signature, err := crypto.Sign(string(testnetKey.Private), forsign)
		if err != nil {
			result.Error = err.Error()
			return result
		}
		toSerialize.BinSignatures = converter.EncodeLengthPlusData(signature)
		toSerialize.PublicKey = pub

		data := make([]byte, 0)
		data = append(append(data, converter.EncodeLength(int64(len(``)))...), []byte(``)...)
		data = append(append(data, converter.EncodeLength(int64(len(pubhex)))...), []byte(pubhex)...)
		toSerialize.Data = converter.EncodeLengthPlusData(data)

		serializedData, err := msgpack.Marshal(toSerialize)
		if err != nil {
			result.Error = err.Error()
			return result
		}
		if _, err = model.SendTx(int64(info.ID), c.SessCitizenID,
			append([]byte{128}, serializedData...)); err != nil {
			result.Error = err.Error()
			return result
		}
		testnetKey.ID = c.SessCitizenID
		testnetKey.StateID = c.SessStateID
		testnetKey.Private = []byte(priv)
		testnetKey.Wallet = idnew
		err = testnetKey.Create()
		if err != nil {
			result.Error = err.Error()
			return result
		}
	}

	result.Generated, err = testnetKey.GetGeneratedCount(c.SessCitizenID, c.SessStateID)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.Available, err = testnetKey.GetAvailableCount(c.SessCitizenID, c.SessStateID)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.Used = result.Generated - result.Available

	return result
}
