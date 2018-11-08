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

package smart

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/GenesisKernel/go-genesis/packages/utils"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"
	"github.com/shopspring/decimal"

	log "github.com/sirupsen/logrus"
)

func logError(err error, errType string, comment string) error {
	log.WithFields(log.Fields{"type": errType, "error": err}).Error(comment)
	return err
}

func logErrorf(pattern string, param interface{}, errType string, comment string) error {
	err := fmt.Errorf(pattern, param)
	log.WithFields(log.Fields{"type": errType, "error": err}).Error(comment)
	return err
}

func logErrorShort(err error, errType string) error {
	return logError(err, errType, err.Error())
}

func logErrorfShort(pattern string, param interface{}, errType string) error {
	return logErrorShort(fmt.Errorf(pattern, param), errType)
}

func logErrorValue(err error, errType string, comment, value string) error {
	log.WithFields(log.Fields{"type": errType, "error": err, "value": value}).Error(comment)
	return err
}

func logErrorDB(err error, comment string) error {
	return logError(err, consts.DBError, comment)
}

func unmarshalJSON(input []byte, v interface{}, comment string) (err error) {
	if err = json.Unmarshal(input, v); err != nil {
		return logErrorValue(err, consts.JSONUnmarshallError, comment, string(input))
	}
	return nil
}

func marshalJSON(v interface{}, comment string) (out []byte, err error) {
	out, err = json.Marshal(v)
	if err != nil {
		logError(err, consts.JSONMarshallError, comment)
	}
	return
}

func validateAccess(funcName string, sc *SmartContract, contracts ...string) error {
	if !accessContracts(sc, contracts...) {
		err := fmt.Errorf(eAccessContract, funcName, strings.Join(contracts, ` or `))
		return logError(err, consts.IncorrectCallingContract, err.Error())
	}
	return nil
}

func CallContract(contractName string, ecosystemID int64, params map[string]string, paramsForSign []string) (*blockchain.Transaction, error) {
	NodePrivateKey, NodePublicKey, err := utils.GetNodeKeys()
	if err != nil {
		return nil, err
	}
	smartTx, err := blockchain.BuildTransaction(blockchain.Transaction{
		Header: blockchain.TxHeader{
			Name:        contractName,
			Time:        time.Now().Unix(),
			EcosystemID: ecosystemID,
			KeyID:       conf.Config.KeyID,
		},
		SignedBy: PubToID(NodePublicKey),
		Params:   params,
	},
		NodePrivateKey,
		NodePublicKey,
		paramsForSign...,
	)
	if err != nil {
		return nil, err
	}
	return smartTx, nil
}

func getFieldDefaultValue(fieldType string) interface{} {
	switch fieldType {
	case "bool":
		return false
	case "float64":
		return float64(0)
	case "int64":
		return int64(0)
	case script.Decimal:
		return decimal.New(0, consts.MoneyDigits)
	case "string":
		return ""
	case "[]uint8":
		return []byte{}
	case "[]interface {}":
		return []interface{}{}
	case script.File:
		return types.NewFile()
	}
	return nil
}

func FillTxData(fieldInfos []*script.FieldInfo, params map[string]string, files map[string]*tx.File, forsign []string) (map[string]interface{}, error) {
	resultParams := map[string]interface{}{}
	for _, fitem := range fieldInfos {
		var err error
		var v interface{}
		var forv string
		var isforv bool
		if _, ok := params[fitem.Name]; !ok && fitem.ContainsTag(script.TagOptional) {
			resultParams[fitem.Name] = getFieldDefaultValue(fitem.Type.String())
			continue
		}
		if _, ok := files[fitem.Name]; !ok && fitem.ContainsTag(script.TagOptional) {
			resultParams[fitem.Name] = getFieldDefaultValue(fitem.Type.String())
			continue
		}
		if fitem.Type.String() == "types.File" {
			file, ok := files[fitem.Name]
			if !ok {
				return nil, nil
			}
			fileMap := map[string]interface{}{}
			fileMap["Body"] = file.Data
			fileMap["MimeType"] = file.MimeType
			fileMap["Name"] = fitem.Name
			resultParams[fitem.Name] = fileMap
			forsign = append(forsign, file.MimeType, file.Hash)
			continue
		}

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
			v, err = hex.DecodeString(params[fitem.Name])
		case `[]interface {}`:
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
		if resultParams[fitem.Name] == nil {
			resultParams[fitem.Name] = v
		}
		if err != nil {
			return nil, err
		}
		if strings.Index(fitem.Tags, `image`) >= 0 {
			continue
		}
		if isforv {
			v = forv
		}
		forsign = append(forsign, fmt.Sprintf("%v", v))
	}
	return resultParams, nil
}
