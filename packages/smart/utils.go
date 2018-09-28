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

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"gopkg.in/vmihailenco/msgpack.v2"
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

func FillTxData(fieldInfos []*script.FieldInfo, input []byte,
	forsign []string) (txData map[string]interface{}, err error) {
	txData = make(map[string]interface{})
	for _, fitem := range fieldInfos {
		var v interface{}
		var forv string
		var isforv, skipFor bool

		if fitem.ContainsTag(script.TagFile) {
			var (
				data []byte
				file *tx.File
			)
			if err = converter.BinUnmarshal(&input, &data); err != nil {
				log.WithFields(log.Fields{"error": err, "type": consts.UnmarshallingError}).Error("bin unmarshalling file")
				return
			}
			if err = msgpack.Unmarshal(data, &file); err != nil {
				log.WithFields(log.Fields{"error": err, "type": consts.UnmarshallingError}).Error("unmarshalling file msgpack")
				return
			}

			txData[fitem.Name] = file.Data
			txData[fitem.Name+"MimeType"] = file.MimeType
			if forsign != nil {
				forsign = append(forsign, file.MimeType, file.Hash)
			}
			continue
		}
		if fitem.ContainsTag(script.TagOptional) && len(input) == 0 {
			switch fitem.Type.String() {
			case `uint64`:
				v = uint64(0)
			case `float64`:
				v = float64(0)
			case `int64`:
				v = int64(0)
			case script.Decimal:
				v = decimal.New(0, 0)
			case `string`, `[]uint8`, `[]interface {}`:
				v = ``
			}
			skipFor = true
		} else {
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
				if err = converter.BinUnmarshal(&input, &s); err != nil {
					log.WithFields(log.Fields{"error": err, "type": consts.UnmarshallingError}).Error("bin unmarshalling script.Decimal")
					return
				}
				v, err = decimal.NewFromString(s)
			case `string`:
				var s string
				if err = converter.BinUnmarshal(&input, &s); err != nil {
					log.WithFields(log.Fields{"error": err, "type": consts.UnmarshallingError}).Error("bin unmarshalling string")
					return
				}
				v = s
			case `[]uint8`:
				var b []byte
				if err = converter.BinUnmarshal(&input, &b); err != nil {
					log.WithFields(log.Fields{"error": err, "type": consts.UnmarshallingError}).Error("bin unmarshalling string")
					return
				}
				v = hex.EncodeToString(b)
			case `[]interface {}`:
				var count int64
				count, err = converter.DecodeLength(&input)
				if err != nil {
					log.WithFields(log.Fields{"error": err, "type": consts.UnmarshallingError}).Error("bin unmarshalling []interface{}")
					return
				}
				isforv = true
				list := make([]interface{}, 0)
				for count > 0 {
					var length int64
					length, err = converter.DecodeLength(&input)
					if err != nil {
						log.WithFields(log.Fields{"error": err, "type": consts.UnmarshallingError}).Error("bin unmarshalling tx length")
						return
					}
					if len(input) < int(length) {
						log.WithFields(log.Fields{"error": err, "type": consts.UnmarshallingError, "length": int(length), "slice length": len(input)}).Error("incorrect tx size")
						err = errInputSlice
						return
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
		}
		if txData[fitem.Name] == nil {
			txData[fitem.Name] = v
		}
		if err != nil {
			return
		}
		if strings.Index(fitem.Tags, `image`) >= 0 {
			continue
		}
		if isforv {
			v = forv
		}
		if forsign != nil && !skipFor {
			forsign = append(forsign, fmt.Sprintf("%v", v))
		}
	}
	if forsign != nil {
		txData[`forsign`] = strings.Join(forsign, ",")
	}
	return
}
