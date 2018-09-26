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
	if len(params) != len(fieldInfos) {
		return nil, fmt.Errorf("Invalid number of parameters")
	}

	txData := make(map[string]interface{})

	for _, fitem := range fieldInfos {
		var (
			v   interface{}
			ok  bool
			err error
		)
		index := fitem.Name
		switch fitem.Type.String() {
		case `bool`:
			if v, ok = params[index].(bool); !ok {
				err = fmt.Errorf("Invalid bool type")
				break
			}
		case `float64`:
			if v, ok = params[index].(float64); !ok {
				err = fmt.Errorf("Invalid float type")
				break
			}
		case `int64`:
			switch t := params[index].(type) {
			case int64:
				v = t
			case uint64:
				v = int64(t)
			default:
				err = fmt.Errorf("Invalid int type")
			}
		case script.Decimal:
			var s string
			if s, ok = params[index].(string); !ok {
				err = fmt.Errorf("Invalid money type")
				break
			}
			v, err = decimal.NewFromString(s)
			if err != nil {
				break
			}
		case `string`:
			if v, ok = params[index].(string); !ok {
				err = fmt.Errorf("Invalid string type")
				break
			}
		case `[]uint8`:
			if v, ok = params[index].([]byte); !ok {
				err = fmt.Errorf("Invalid bytes type")
				break
			}
		case `[]interface {}`:
			if v, ok = params[index].([]interface{}); !ok {
				err = fmt.Errorf("Invalid array type")
				break
			}
		case script.File:
			var val map[interface{}]interface{}
			if val, ok = params[index].(map[interface{}]interface{}); !ok {
				err = fmt.Errorf("Invalid file type")
				break
			}

			if v, ok = types.NewFileFromMap(val); !ok {
				err = fmt.Errorf("Invalid attrs of file")
				break
			}
		}
		if err != nil {
			return nil, fmt.Errorf("Invalid param '%s': %s", index, err)
		}

		if _, ok = txData[fitem.Name]; !ok {
			txData[fitem.Name] = v
		}
	}

	return txData, nil
}
