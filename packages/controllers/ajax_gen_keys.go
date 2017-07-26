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
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
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

	count := converter.StrToInt64(c.r.FormValue("count"))
	if count < 1 || count > 50 {
		result.Error = `Count must be from 1 to 50`
		return result
	}
	stateParameters := &model.StateParameters{}
	stateParameters.SetTableName(c.SessStateID)
	err = stateParameters.GetByName("gov_account")
	if err != nil {
		result.Error = err.Error()
		return result
	}
	govAccount := stateParameters.Value

	if c.SessCitizenID != converter.StrToInt64(govAccount) || len(govAccount) == 0 {
		result.Error = `Access denied`
		return result
	}

	testnetKey := &model.TestnetKeys{}
	err = testnetKey.GetByWallet(c.SessCitizenID)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if len(testnetKey.Private) == 0 {
		result.Error = `unknown private key`
		return result
	}
	//	bkey, err := hex.DecodeString(privKey)
	//	pubkey := lib.PrivateToPublic(bkey)

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

		wallet := &model.DltWallets{}
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
		var flags uint8

		ctime := uint32(time.Now().Unix())
		info := (*contract).Block.Info.(*script.ContractInfo)
		forsign := fmt.Sprintf("%d,%d,%d,%d,%d", info.ID, ctime, uint64(c.SessCitizenID), c.SessStateID, flags)
		pubhex := hex.EncodeToString(pub)
		forsign += fmt.Sprintf(",%v,%v", ``, pubhex)
		signature, err := crypto.Sign(string(testnetKey.Private), forsign)
		if err != nil {
			result.Error = err.Error()
			return result
		}

		sign := make([]byte, 0)
		converter.EncodeLenByte(&sign, signature)
		data := make([]byte, 0)
		header := consts.TXHeader{
			Type:     int32(contract.Block.Info.(*script.ContractInfo).ID),
			Time:     uint32(ctime),
			WalletID: uint64(c.SessCitizenID),
			StateID:  int32(c.SessStateID),
			Flags:    flags,
			Sign:     sign,
		}
		_, err = converter.BinMarshal(&data, &header)
		if err != nil {
			result.Error = err.Error()
			return result
		}
		data = append(append(data, converter.EncodeLength(int64(len(``)))...), []byte(``)...)
		data = append(append(data, converter.EncodeLength(int64(len(pubhex)))...), []byte(pubhex)...)
		err = c.SendTx(int64(header.Type), c.SessCitizenID, data)
		if err != nil {
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
