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
	"encoding/json"
	"net/http"
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

type columnInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Perm string `json:"perm"`
}

type tableResult struct {
	Name       string       `json:"name"`
	Insert     string       `json:"insert"`
	NewColumn  string       `json:"new_column"`
	Update     string       `json:"update"`
	Read       string       `json:"read,omitempty"`
	Filter     string       `json:"filter,omitempty"`
	Conditions string       `json:"conditions"`
	AppID      string       `json:"app_id"`
	Columns    []columnInfo `json:"columns"`
}

func table(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	var result tableResult

	prefix := getPrefix(data)
	table := &model.Table{}
	table.SetTablePrefix(prefix)
	_, err = table.Get(nil, strings.ToLower(data.params[`name`].(string)))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting table")
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}

	if len(table.Name) > 0 {
		var perm map[string]string
		err := json.Unmarshal([]byte(table.Permissions), &perm)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("Unmarshalling table permissions to json")
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}
		var cols map[string]string
		err = json.Unmarshal([]byte(table.Columns), &cols)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("Unmarshalling table columns to json")
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}
		columns := make([]columnInfo, 0)
		for key, value := range cols {
			colType, err := model.GetColumnType(prefix+`_`+data.params[`name`].(string), key)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting column type from db")
				return errorAPI(w, err.Error(), http.StatusInternalServerError)
			}
			columns = append(columns, columnInfo{Name: key, Perm: value,
				Type: colType})
		}
		result = tableResult{
			Name:       table.Name,
			Insert:     perm[`insert`],
			NewColumn:  perm[`new_column`],
			Update:     perm[`update`],
			Read:       perm[`read`],
			Filter:     perm[`filter`],
			Conditions: table.Conditions,
			AppID:      converter.Int64ToStr(table.AppID),
			Columns:    columns,
		}
	} else {
		return errorAPI(w, `E_TABLENOTFOUND`, http.StatusBadRequest, data.params[`name`].(string))
	}
	data.result = &result
	return
}
