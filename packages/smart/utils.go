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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/AplaProject/go-apla/packages/blockchain"
	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/types"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/AplaProject/go-apla/packages/utils/tx"
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

func FillTxData(fieldInfos []*script.FieldInfo, params map[string]string, files map[string]*tx.File, forsign []string) (map[string]interface{}, error) {
	resultParams := map[string]interface{}{}
	for _, fitem := range fieldInfos {
		var err error
		var v interface{}
		var forv string
		var isforv bool
		if _, ok := files[fitem.Name]; !ok && fitem.ContainsTag(script.TagOptional) && len(params[fitem.Name]) == 0 {
			resultParams[fitem.Name] = getFieldDefaultValue(fitem.Original)
			continue
		}
		if fitem.Type.String() == script.File {
			file, ok := files[fitem.Name]
			if !ok {
				return nil, nil
			}
			fileMap := types.LoadMap(map[string]interface{}{
				"Body":     file.Data,
				"MimeType": file.MimeType,
				"Name":     fitem.Name})
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
