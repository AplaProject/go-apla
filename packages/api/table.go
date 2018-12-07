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

package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	"github.com/gorilla/mux"
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

func getTableHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	logger := getLogger(r)
	client := getClient(r)
	prefix := client.Prefix()

	table := &model.Table{}
	table.SetTablePrefix(prefix)

	_, err := table.Get(nil, strings.ToLower(params["name"]))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting table")
		errorResponse(w, err)
		return
	}

	if len(table.Name) == 0 {
		errorResponse(w, errTableNotFound.Errorf(params["name"]))
		return
	}

	var columnsMap map[string]string
	err = json.Unmarshal([]byte(table.Columns), &columnsMap)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("Unmarshalling table columns to json")
		errorResponse(w, err)
		return
	}

	columns := make([]columnInfo, 0)
	for key, value := range columnsMap {
		colType, err := model.GetColumnType(prefix+`_`+params["name"], key)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting column type from db")
			errorResponse(w, err)
			return
		}
		columns = append(columns, columnInfo{
			Name: key,
			Perm: value,
			Type: colType,
		})
	}

	jsonResponse(w, &tableResult{
		Name:       table.Name,
		Insert:     table.Permissions.Insert,
		NewColumn:  table.Permissions.NewColumn,
		Update:     table.Permissions.Update,
		Read:       table.Permissions.Read,
		Filter:     table.Permissions.Filter,
		Conditions: table.Conditions,
		AppID:      converter.Int64ToStr(table.AppID),
		Columns:    columns,
	})
}
