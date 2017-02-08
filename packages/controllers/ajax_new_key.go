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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const ANewKey = `ajax_new_key`

type NewKey struct {
	//	Address string `json:"address"`
	Private string `json:"private"`
	Seed    string `json:"seed"`
	Error   string `json:"error"`
}

var words []string

func init() {
	newPage(ANewKey, `json`)
}

func (c *Controller) AjaxNewKey() interface{} {
	var result NewKey

	if len(words) == 0 {
		in, _ := ioutil.ReadFile(*utils.Dir + `/words.txt`)
		if len(in) > 0 {
			list := strings.Replace(strings.Replace(string(in), "\r", "", -1), "\n", ` `, -1)
			tmp := strings.Split(strings.Replace(strings.Replace(list, `"`, "", -1), ",", ` `, -1), ` `)
			for _, v := range tmp {
				if v = strings.TrimSpace(v); len(v) > 0 {
					words = append(words, v)
				}
			}
		}
		//		fmt.Println(`Words`, words)
	}
	var seed string
	key := c.r.FormValue("key")
	name := c.r.FormValue("name")
	stateId := utils.StrToInt64(c.r.FormValue("state_id"))
	bkey, err := hex.DecodeString(key)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if stateId == 0 {
		result.Error = `state_id has not been specified`
		return result
	}
	pubkey := lib.PrivateToPublic(bkey)
	idkey := int64(lib.Address(pubkey))
	gov_account, _ := utils.StateParam(stateId, `gov_account`)
	if len(gov_account) == 0 {
		result.Error = `unknown gov_account`
		return result
	}
	if utils.StrToInt64(gov_account) != idkey {
		result.Error = `access denied`
		return result
	}
	var priv []byte
	if len(words) == 0 {
		spriv, _ := lib.GenKeys()
		priv, _ = hex.DecodeString(spriv)
	} else {
		phrase := make([]string, 0)
		rand.Seed(int64(lib.Time32()))
		for len(phrase) < 15 {
			rnd := rand.Intn(len(words))
			if len(words[rnd]) > 0 {
				phrase = append(phrase, strings.ToLower(words[rnd]))
			}
		}
		seed = strings.Join(phrase, ` `)
		sha := sha256.Sum256([]byte(seed))
		priv = sha[:]
	}
	if len(priv) != 32 {
		result.Error = `wrong private key`
		return result
	}
	pub := lib.PrivateToPublic(priv)
	idnew := int64(lib.Address(pub))

	exist, err := c.Single(`select wallet_id from dlt_wallets where wallet_id=?`, idnew).Int64()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if exist != 0 {
		result.Error = `key already exists`
		return result
	}
	contract := smart.GetContract(`GenCitizen`, uint32(stateId))
	if contract == nil {
		result.Error = `GenCitizen contract has not been found`
		return result
	}
	var flags uint8

	ctime := lib.Time32()
	info := (*contract).Block.Info.(*script.ContractInfo)
	forsign := fmt.Sprintf("%d,%d,%d,%d,%d", info.Id, ctime, uint64(idkey), stateId, flags)
	pubhex := hex.EncodeToString(pub)
	forsign += fmt.Sprintf(",%v,%v", name, pubhex)

	signature, err := lib.SignECDSA(key, forsign)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	sign := make([]byte, 0)
	lib.EncodeLenByte(&sign, signature)
	data := make([]byte, 0)
	header := consts.TXHeader{
		Type:     int32(contract.Block.Info.(*script.ContractInfo).Id),
		Time:     uint32(ctime),
		WalletId: uint64(idkey),
		StateId:  int32(stateId),
		Flags:    flags,
		Sign:     sign,
	}
	_, err = lib.BinMarshal(&data, &header)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	data = append(append(data, lib.EncodeLength(int64(len(name)))...), []byte(name)...)
	data = append(append(data, lib.EncodeLength(int64(len(pubhex)))...), []byte(pubhex)...)

	/*	fmt.Printf("NewKey For %s %d\r\n", forsign, len(forsign))
		fmt.Printf("NewKey Sign %x %d\r\n", sign, len(sign))
		fmt.Printf("NewKey Key %x %d\r\n", pubkey, len(pubkey))
	*/
	md5 := utils.Md5(data)
	err = c.ExecSql(`INSERT INTO transactions_status (
			hash, time,	type, wallet_id, citizen_id	) VALUES (
			[hex], ?, ?, ?, ? )`, md5, time.Now().Unix(), header.Type, int64(idkey), int64(idkey))
	if err != nil {
		result.Error = err.Error()
		return result
	}
	err = c.ExecSql("INSERT INTO queue_tx (hash, data) VALUES ([hex], [hex])", md5, hex.EncodeToString(data))
	if err != nil {
		result.Error = err.Error()
		return result
	}

	result.Seed = seed
	result.Private = hex.EncodeToString(priv)
	return result
}
