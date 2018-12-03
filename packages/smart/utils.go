// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package smart

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/types"
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
	if conf.Config.FuncBench {
		return nil
	}

	if !accessContracts(sc, contracts...) {
		err := fmt.Errorf(eAccessContract, funcName, strings.Join(contracts, ` or `))
		return logError(err, consts.IncorrectCallingContract, err.Error())
	}
	return nil
}

func FillTxData(fieldInfos []*script.FieldInfo, params map[string]interface{}) (map[string]interface{}, error) {
	txData := make(map[string]interface{})
	for _, fitem := range fieldInfos {
		var (
			v     interface{}
			ok    bool
			err   error
			index = fitem.Name
		)

		if _, ok := params[index]; !ok {
			if fitem.ContainsTag(script.TagOptional) {
				txData[index] = getFieldDefaultValue(fitem.Original)
				continue
			}
			return nil, fmt.Errorf(eParamNotFound, index)
		}

		switch fitem.Original {
		case script.DtBool:
			if v, ok = params[index].(bool); !ok {
				err = fmt.Errorf("Invalid bool type")
				break
			}
		case script.DtFloat:
			switch val := params[index].(type) {
			case float64:
				v = val
			case uint64:
				v = float64(val)
			case int64:
				v = float64(val)
			default:
				err = fmt.Errorf("Invalid float type")
				break
			}
		case script.DtInt, script.DtAddress:
			switch t := params[index].(type) {
			case int64:
				v = t
			case uint64:
				v = int64(t)
			default:
				err = fmt.Errorf("Invalid int type")
			}
		case script.DtMoney:
			var s string
			if s, ok = params[index].(string); !ok {
				err = fmt.Errorf("Invalid money type")
				break
			}
			v, err = decimal.NewFromString(s)
			if err != nil {
				break
			}
		case script.DtString:
			if v, ok = params[index].(string); !ok {
				err = fmt.Errorf("Invalid string type")
				break
			}
		case script.DtBytes:
			if v, ok = params[index].([]byte); !ok {
				err = fmt.Errorf("Invalid bytes type")
				break
			}
		case script.DtArray:
			if v, ok = params[index].([]interface{}); !ok {
				err = fmt.Errorf("Invalid array type")
				break
			}
		case script.DtMap:
			var val map[interface{}]interface{}
			if val, ok = params[index].(map[interface{}]interface{}); !ok {
				err = fmt.Errorf("Invalid map type")
				break
			}
			vMap := types.NewMap()
			for key, item := range val {
				vMap.Set(fmt.Sprint(key), item)
			}
			v = vMap
		case script.DtFile:
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

	if len(txData) != len(fieldInfos) {
		return nil, fmt.Errorf("Invalid number of parameters")
	}

	return txData, nil
}

func getFieldDefaultValue(fieldType uint32) interface{} {
	switch fieldType {
	case script.DtBool:
		return false
	case script.DtFloat:
		return float64(0)
	case script.DtInt, script.DtAddress:
		return int64(0)
	case script.DtMoney:
		return decimal.New(0, consts.MoneyDigits)
	case script.DtString:
		return ""
	case script.DtBytes:
		return []byte{}
	case script.DtArray:
		return []interface{}{}
	case script.DtMap:
		return types.NewMap()
	case script.DtFile:
		return types.NewFile()
	}
	return nil
}
