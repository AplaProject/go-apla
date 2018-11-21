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
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/smart"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type vdeCreateResult struct {
	Result bool `json:"result"`
}

func vdeCreate(w http.ResponseWriter, r *http.Request) {
	client := getClient(r)
	logger := getLogger(r)

	if model.IsTable(fmt.Sprintf(`%d_vde_tables`, client.EcosystemID)) {
		errorResponse(w, errVDECreated)
		return
	}
	sp := &model.StateParameter{}
	sp.SetTablePrefix(converter.Int64ToStr(client.EcosystemID))
	if _, err := sp.Get(nil, `founder_account`); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating vde")
		errorResponse(w, err)
		return
	}
	if converter.StrToInt64(sp.Value) != client.KeyID {
		logger.WithFields(log.Fields{"type": consts.AccessDenied, "error": fmt.Errorf(`Access denied`)}).Error("creating vde")
		errorResponse(w, errPermission)
		return
	}
	if err := model.ExecSchemaLocalData(int(client.EcosystemID), client.KeyID); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating vde")
		errorResponse(w, err)
		return
	}
	smart.LoadVDEContracts(nil, converter.Int64ToStr(client.EcosystemID))

	jsonResponse(w, &vdeCreateResult{Result: true})
}

// InitSmartContract is initializes smart contract
func InitSmartContract(sc *smart.SmartContract, data []byte) error {
	if err := msgpack.Unmarshal(data, &sc.TxSmart); err != nil {
		return err
	}

	sc.TxContract = smart.VMGetContract(smart.GetVM(), sc.TxSmart.Header.Name, uint32(sc.TxSmart.Header.EcosystemID))
	if sc.TxContract == nil {
		return fmt.Errorf(`unknown contract %s`, sc.TxSmart.Header.Name)
	}
	forsign := ""

	params := sc.TxSmart.Params
	sc.TxData = make(map[string]interface{})

	if sc.TxContract.Block.Info.(*script.ContractInfo).Tx != nil {
		for _, fitem := range *sc.TxContract.Block.Info.(*script.ContractInfo).Tx {
			var v interface{}
			var err error
			var forv string
			var isforv bool
			switch fitem.Type.String() {
			case `uint64`:
				var val uint64
				val = converter.StrToUint64(params[fitem.Name])
				v = val
			case `float64`:
				var val float64
				val = converter.StrToFloat64(params[fitem.Name])
				v = val
			case `int64`:
				v = converter.StrToInt64(params[fitem.Name])
			case script.Decimal:
				v, err = decimal.NewFromString(params[fitem.Name])
			case `string`:
				v = params[fitem.Name]
			case `[]uint8`:
				v, _ = hex.DecodeString(params[fitem.Name])
			case `[]interface {}`:
				isforv = true
				var list []string
				for key, value := range params {
					if key == fitem.Name+`[]` && len(value) > 0 {
						count := converter.StrToInt(value)
						for i := 0; i < count; i++ {
							list = append(list, params[fmt.Sprintf(`%s[%d]`, fitem.Name, i)])
						}
					}
				}
				if len(list) > 0 {
					forv = strings.Join(list, `,`)
				}
				v = list
			}
			if sc.TxData[fitem.Name] == nil {
				sc.TxData[fitem.Name] = v
			}
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

	sc := smart.SmartContract{VDE: true, TxHash: hash, Rand: rand.New(rand.NewSource(time.Now().Unix()))}
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
