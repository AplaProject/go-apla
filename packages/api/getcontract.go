// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.
//
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
//
// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package api

import (
	"net/http"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/smart"

	log "github.com/sirupsen/logrus"
)

type contractField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Optional bool   `json:"optional"`
}

type getContractResult struct {
	ID       uint32          `json:"id"`
	StateID  uint32          `json:"state"`
	Active   bool            `json:"active"`
	TableID  string          `json:"tableid"`
	WalletID string          `json:"walletid"`
	TokenID  string          `json:"tokenid"`
	Address  string          `json:"address"`
	Fields   []contractField `json:"fields"`
	Name     string          `json:"name"`
}

func getContract(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var result getContractResult

	cntname := data.params[`name`].(string)
	contract := smart.VMGetContract(data.vm, cntname, uint32(data.ecosystemId))
	if contract == nil {
		logger.WithFields(log.Fields{"type": consts.ContractError, "contract_name": cntname}).Error("contract name")
		return errorAPI(w, `E_CONTRACT`, http.StatusBadRequest, cntname)
	}
	info := (*contract).Block.Info.(*script.ContractInfo)
	fields := make([]contractField, 0)
	result = getContractResult{
		ID:   uint32(info.Owner.TableID + consts.ShiftContractID),
		Name: info.Name, StateID: info.Owner.StateID,
		Active: info.Owner.Active, TableID: converter.Int64ToStr(info.Owner.TableID),
		WalletID: converter.Int64ToStr(info.Owner.WalletID),
		TokenID:  converter.Int64ToStr(info.Owner.TokenID),
		Address:  converter.AddressToString(info.Owner.WalletID),
	}

	if info.Tx != nil {
		for _, fitem := range *info.Tx {
			fields = append(fields, contractField{
				Name:     fitem.Name,
				Type:     getFieldTypeAlias(fitem.Type.String()),
				Optional: fitem.ContainsTag(script.TagOptional),
			})
		}
	}
	result.Fields = fields

	data.result = result
	return nil
}

func getFieldTypeAlias(t string) string {
	var fieldTypeAliases = map[string]string{
		"int64":           "int",
		"float64":         "float",
		"decimal.Decimal": "money",
		"[]uint8":         "bytes",
		"[]interface {}":  "array",
		"*types.Map":      "file",
	}

	if v, ok := fieldTypeAliases[t]; ok {
		return v
	}
	return t
}
