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
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

type rowResult struct {
	Value map[string]string `json:"value"`
}

func row(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	cols := `*`
	if len(data.params[`columns`].(string)) > 0 {
		cols = converter.EscapeName(data.params[`columns`].(string))
	}
	var table string
	name := data.params[`name`].(string)
	if model.FirstEcosystemTables[name] {
		table = `1_` + name
	} else {
		table = converter.ParseTable(name, data.ecosystemId)
	}
	row, err := model.GetOneRow(`SELECT `+cols+` FROM "`+table+`" WHERE id = ?`, data.params[`id`].(string)).String()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "table": name, "id": data.params["id"].(string)}).Error("getting one row")
		return errorAPI(w, `E_QUERY`, http.StatusInternalServerError)
	}

	data.result = &rowResult{Value: row}
	return
}
