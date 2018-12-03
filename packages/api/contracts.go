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
	"fmt"
	"net/http"
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/script"

	log "github.com/sirupsen/logrus"
)

type contractsResult struct {
	Count string              `json:"count"`
	List  []map[string]string `json:"list"`
}

func getContracts(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	var limit int

	table := `1_contracts`

	where := fmt.Sprintf(`ecosystem='%d'`, data.ecosystemId)

	count, err := model.GetRecordsCountTx(nil, table, where)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "table": table}).Error("Getting table records count")
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}

	if data.params[`limit`].(int64) > 0 {
		limit = int(data.params[`limit`].(int64))
	} else {
		limit = 25
	}
	list, err := model.GetAll(fmt.Sprintf(`select * from "%s" where %s order by id desc offset %d `, table, where, data.params[`offset`].(int64)), limit)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting all")
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	for ind, val := range list {
		if val[`wallet_id`] == `NULL` {
			list[ind][`wallet_id`] = ``
			list[ind][`address`] = ``
		} else {
			list[ind][`address`] = converter.AddressToString(converter.StrToInt64(val[`wallet_id`]))
		}
		if val[`active`] == `NULL` {
			list[ind][`active`] = ``
		}
		cntlist, err := script.ContractsList(val[`value`])
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.ContractError, "error": err}).Error("getting contract list")
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}
		list[ind][`name`] = strings.Join(cntlist, `,`)
	}
	data.result = &listResult{
		Count: converter.Int64ToStr(count), List: list,
	}
	return
}
