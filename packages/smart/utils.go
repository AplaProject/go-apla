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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/types"
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

func FillTxData(fieldInfos []*script.FieldInfo, params map[string]interface{}) (map[string]interface{}, error) {
	// TODO: remove forsign

	if len(params) != len(fieldInfos) {
		return nil, fmt.Errorf("Invalid number of parameters")
	}

	txData := make(map[string]interface{})

	for _, fitem := range fieldInfos {
		var err error
		var v interface{}
		var ok bool
		var forv string
		var isforv bool

		index := fitem.Name
		switch fitem.Type.String() {
		case `bool`:
			if v, ok = params[index].(bool); !ok {
				return nil, fmt.Errorf("Incorrect type bool")
			}
		case `uint64`:
			if v, ok = params[index].(uint64); !ok {
				return nil, fmt.Errorf("Incorrect type uint64")
			}
		case `float64`:
			if v, ok = params[index].(float64); !ok {
				return nil, fmt.Errorf("Incorrect type float64")
			}
		case `int64`:
			if v, ok = params[index].(int64); !ok {
				return nil, fmt.Errorf("Incorrect type int64")
			}
		case script.Decimal:
			var s string
			if s, ok = params[index].(string); !ok {
				return nil, fmt.Errorf("Incorrect type money")
			}
			v, err = decimal.NewFromString(s)
			if err != nil {
				return nil, err
			}
		case `string`:
			if v, ok = params[index].(string); !ok {
				return nil, fmt.Errorf("Incorrect type string")
			}
		case `[]uint8`:
			var val []byte
			if val, ok = params[index].([]byte); !ok {
				return nil, fmt.Errorf("Incorrect type []uint8")
			}

			if forv, err = crypto.HashHex(val); err != nil {
				return nil, err
			}

			isforv = true
			v = val
		case `[]interface {}`:
			var val []interface{}
			if val, ok = params[index].([]interface{}); !ok {
				return nil, fmt.Errorf("Incorrect type []interface {}")
			}

			list := make([]string, len(val)+1)
			list[0] = converter.IntToStr(len(val))
			for i, _ := range val {
				list[i+1] = fmt.Sprintf("%v", val[i])
			}

			v = val
			isforv = true
			forv = strings.Join(list, ",")
		case script.File:
			var val map[interface{}]interface{}
			if val, ok = params[index].(map[interface{}]interface{}); !ok {
				return nil, fmt.Errorf("Incorrect type file")
			}

			file := types.File{
				"Name":     val["Name"].(string),
				"MimeType": val["MimeType"].(string),
				"Body":     val["Body"].([]byte),
			}

			v = file
			isforv = true
			forv = "file"
		}

		if _, ok = txData[fitem.Name]; !ok {
			txData[fitem.Name] = v
		}
		if err != nil {
			return nil, err
		}
		if isforv {
			v = forv
		}
	}

	return txData, nil
}
