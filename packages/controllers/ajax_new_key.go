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
	"strconv"
	"strings"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"

	msgpack "gopkg.in/vmihailenco/msgpack.v2"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"
)

const aNewKey = `ajax_new_key`

// NewKey is a structure for the answer of ajax_new_key ajax request
type NewKey struct {
	//	Address string `json:"address"`
	Private string `json:"private"`
	Seed    string `json:"seed"`
	Error   string `json:"error"`
}

var words []string

func init() {
	newPage(aNewKey, `json`)
}

// AjaxNewKey is a controller of ajax_new_key request
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
	stateID, err := strconv.ParseInt(c.r.FormValue("state_id"), 10, 64)
	if err != nil {
		logger.LogInfo(consts.StrtoInt64Error, c.r.FormValue("state_id"))
	}
	bkey, err := hex.DecodeString(key)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if stateID == 0 {
		result.Error = `state_id has not been specified`
		return result
	}
	pubkey, err := crypto.PrivateToPublic(bkey)
	if err != nil {
		log.Fatal(err)
	}
	idkey := crypto.Address(pubkey)
	stateParameter := &model.StateParameter{}
	stateParameter.GetByName("govAccount")
	if len(stateParameter.Value) == 0 {
		result.Error = `unknown govAccount`
		return result
	}

	govAccount, err := strconv.ParseInt(stateParameter.Value, 10, 64)
	if err != nil {
		logger.LogInfo(consts.StrtoInt64Error, stateParameter.Value)
	}

	if govAccount != idkey {
		result.Error = `access denied`
		return result
	}
	var priv []byte
	if len(words) == 0 {
		spriv, _, _ := crypto.GenHexKeys()
		priv, _ = hex.DecodeString(spriv)
	} else {
		phrase := make([]string, 0)
		rand.Seed(time.Now().Unix())
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
	pub, err := crypto.PrivateToPublic(priv)
	if err != nil {
		log.Fatal(err)
	}
	idnew := crypto.Address(pub)

	wallet := &model.DltWallet{}
	wallet.WalletID = idnew
	exist, err := wallet.IsExists()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if exist != false {
		result.Error = `key already exists`
		return result
	}
	contract := smart.GetContract(`GenCitizen`, uint32(stateID))
	if contract == nil {
		result.Error = `GenCitizen contract has not been found`
		return result
	}

	ctime := time.Now().Unix()
	info := (*contract).Block.Info.(*script.ContractInfo)
	toSerialize := tx.SmartContract{
		Header: tx.Header{Type: int(info.ID), Time: ctime,
			UserID: c.SessCitizenID, StateID: c.SessStateID}}
	pubhex := hex.EncodeToString(pub)
	forsign := toSerialize.ForSign() + fmt.Sprintf("%v,%v", name, pubhex)
	signature, err := crypto.Sign(key, forsign)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	toSerialize.BinSignatures = converter.EncodeLengthPlusData(signature)
	toSerialize.PublicKey = pub

	data := make([]byte, 0)
	data = append(append(data, converter.EncodeLength(int64(len(name)))...), []byte(name)...)
	data = append(append(data, converter.EncodeLength(int64(len(pubhex)))...), []byte(pubhex)...)
	toSerialize.Data = converter.EncodeLengthPlusData(data)

	serializedData, err := msgpack.Marshal(toSerialize)
	hash, err := crypto.Hash(serializedData)
	if err != nil {
		log.Fatal(err)
	}
	transactionStatus := &model.TransactionStatus{Hash: hash, Time: time.Now().Unix(), Type: int64(info.ID),
		WalletID: int64(idkey), CitizenID: int64(idkey)}
	err = transactionStatus.Create()
	if err != nil {
		result.Error = err.Error()
		return result
	}

	queueTx := &model.QueueTx{Hash: hash, Data: serializedData} // TODO: serializedData ????
	if err = queueTx.Create(); err != nil {
		result.Error = err.Error()
		return result
	}

	result.Seed = seed
	result.Private = hex.EncodeToString(priv)
	return result
}
