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

package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/smart"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type vdeCreateResult struct {
	Result bool `json:"result"`
}

// InitSmartContract is initializes smart contract
func InitSmartContract(sc *smart.SmartContract, data []byte) error {
	if err := msgpack.Unmarshal(data, &sc.TxSmart); err != nil {
		return err
	}

	sc.TxContract = smart.VMGetContractByID(smart.GetVM(), int32(sc.TxSmart.ID))
	if sc.TxContract == nil {
		return fmt.Errorf(`unknown contract %d`, sc.TxSmart.ID)
	}
	forsign := ""

	input := data[:]
	sc.TxData = make(map[string]interface{})

	if sc.TxContract.Block.Info.(*script.ContractInfo).Tx != nil {
		for _, fitem := range *sc.TxContract.Block.Info.(*script.ContractInfo).Tx {
			var err error
			var v interface{}
			var forv string
			var isforv bool
			switch fitem.Type.String() {
			case `uint64`:
				var val uint64
				converter.BinUnmarshal(&input, &val)
				v = val
			case `float64`:
				var val float64
				converter.BinUnmarshal(&input, &val)
				v = val
			case `int64`:
				v, err = converter.DecodeLenInt64(&input)
			case script.Decimal:
				var s string
				if err := converter.BinUnmarshal(&input, &s); err != nil {
					return err
				}
				v, err = decimal.NewFromString(s)
			case `string`:
				var s string
				if err := converter.BinUnmarshal(&input, &s); err != nil {
					return err
				}
				v = s
			case `[]uint8`:
				var b []byte
				if err := converter.BinUnmarshal(&input, &b); err != nil {
					return err
				}
				v = hex.EncodeToString(b)
			case `[]interface {}`:
				count, err := converter.DecodeLength(&input)
				if err != nil {
					return err
				}
				isforv = true
				list := make([]interface{}, 0)
				for count > 0 {
					length, err := converter.DecodeLength(&input)
					if err != nil {
						return err
					}
					if len(input) < int(length) {
						return fmt.Errorf(`input slice is short`)
					}
					list = append(list, string(input[:length]))
					input = input[length:]
					count--
				}
				if len(list) > 0 {
					slist := make([]string, len(list))
					for j, lval := range list {
						slist[j] = lval.(string)
					}
					forv = strings.Join(slist, `,`)
				}
				v = list
			}
			sc.TxData[fitem.Name] = v
			if err != nil {
				return err
			}
			if strings.Index(fitem.Tags, `image`) >= 0 {
				continue
			}
			if isforv {
				v = forv
			}
			forsign += fmt.Sprintf(",%v", v)
		}
	}
	sc.TxData[`forsign`] = forsign
	return nil
}

// VDEContract is init VDE contract
func VDEContract(r *http.Request, contractData []byte) (result *contractResult, err error) {
	var ret string
	hash, err := crypto.Hash(contractData)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("getting hash of contract data")
		return
	}
	result = &contractResult{Hash: hex.EncodeToString(hash)}

	sc := smart.SmartContract{
		VDE:    true,
		TxHash: hash,
		Rand:   rand.New(rand.NewSource(time.Now().Unix())),
	}

	err = InitSmartContract(&sc, contractData)
	if err != nil {
		result.Message = &txstatusError{Type: "panic", Error: err.Error()}
		return
	}

	token := getToken(r)

	if token.Valid {
		if auth, err := token.SignedString([]byte(jwtSecret)); err == nil {
			sc.TxData[`auth_token`] = auth
		}
	}

	if ret, err = sc.CallContract(); err == nil {
		result.Result = ret
	} else {
		if errResult := json.Unmarshal([]byte(err.Error()), &result.Message); errResult != nil {
			log.WithFields(log.Fields{
				"type":  consts.JSONUnmarshallError,
				"text":  err.Error(),
				"error": errResult}).Error("unmarshalling contract error")

			result.Message = &txstatusError{Type: "panic", Error: errResult.Error()}
		}
	}
	return
}
