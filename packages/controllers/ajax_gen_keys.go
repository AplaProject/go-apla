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

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
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

	count := utils.StrToInt64(c.r.FormValue("count"))
	if count < 1 || count > 50 {
		result.Error = `Count must be from 1 to 50`
		return result
	}
	govAccount, err := utils.StateParam(int64(c.SessStateID), `gov_account`)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if c.SessCitizenID != utils.StrToInt64(govAccount) || len(govAccount) == 0 {
		result.Error = `Access denied`
		return result
	}
	privKey, err := c.Single(`select private from testnet_emails where wallet=?`, c.SessCitizenID).String()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if len(privKey) == 0 {
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
		spriv, _, _ := lib.GenHexKeys()
		priv, _ = hex.DecodeString(spriv)

		pub := lib.PrivateToPublic(priv)
		idnew := int64(lib.Address(pub))

		exist, err := c.Single(`select wallet_id from dlt_wallets where wallet_id=?`, idnew).Int64()
		if err != nil {
			result.Error = err.Error()
			return result
		}
		if exist != 0 {
			i--
			continue
		}
		var flags uint8

		ctime := lib.Time32()
		info := (*contract).Block.Info.(*script.ContractInfo)
		forsign := fmt.Sprintf("%d,%d,%d,%d,%d", info.ID, ctime, uint64(c.SessCitizenID), c.SessStateID, flags)
		pubhex := hex.EncodeToString(pub)
		forsign += fmt.Sprintf(",%v,%v", ``, pubhex)
		signature, err := lib.SignECDSA(privKey, forsign)
		if err != nil {
			result.Error = err.Error()
			return result
		}

		sign := make([]byte, 0)
		lib.EncodeLenByte(&sign, signature)
		data := make([]byte, 0)
		header := consts.TXHeader{
			Type:     int32(contract.Block.Info.(*script.ContractInfo).ID),
			Time:     uint32(ctime),
			WalletID: uint64(c.SessCitizenID),
			StateID:  int32(c.SessStateID),
			Flags:    flags,
			Sign:     sign,
		}
		_, err = lib.BinMarshal(&data, &header)
		if err != nil {
			result.Error = err.Error()
			return result
		}
		data = append(append(data, lib.EncodeLength(int64(len(``)))...), []byte(``)...)
		data = append(append(data, lib.EncodeLength(int64(len(pubhex)))...), []byte(pubhex)...)
		err = c.SendTx(int64(header.Type), c.SessCitizenID, data)
		if err != nil {
			result.Error = err.Error()
			return result
		}
		err = c.ExecSQL(`insert into testnet_keys (id, state_id, private, wallet) values(?,?,?,?)`,
			c.SessCitizenID, c.SessStateID, spriv, idnew)
		if err != nil {
			result.Error = err.Error()
			return result
		}
	}

	result.Generated, err = c.Single(`select count(id) from testnet_keys where id=? and state_id=?`, c.SessCitizenID, c.SessStateID).Int64()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.Available, err = c.Single(`select count(id) from testnet_keys where id=? and state_id=? and status=0`, c.SessCitizenID, c.SessStateID).Int64()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.Used = result.Generated - result.Available
	//	result.Generated = generated, Available: available, Used: generated - available}

	return result
}
