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
	"net/http"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/utils/tx"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type listResult struct {
	Count string              `json:"count"`
	List  []map[string]string `json:"list"`
}

type listForm struct {
	paginatorForm
	rowForm
}

func (f *listForm) Validate(r *http.Request) error {
	if err := f.paginatorForm.Validate(r); err != nil {
		return err
	}
	return f.rowForm.Validate(r)
}

func getListHandler(w http.ResponseWriter, r *http.Request) {
	form := &listForm{}
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)
	client := getClient(r)
	logger := getLogger(r)

	sc := smart.SmartContract{
		OBS: conf.Config.IsSupportingOBS(),
		VM:  smart.GetVM(),
		TxSmart: tx.SmartContract{
			Header: tx.Header{
				EcosystemID: client.EcosystemID,
				KeyID:       client.KeyID,
				NetworkID:   consts.NETWORK_ID,
			},
		},
	}

	table := params["name"]
	q := model.GetTableQuery(table, client.EcosystemID)

	table = converter.ParseTable(table, client.EcosystemID)
	var err error
	_, form.Columns, err = sc.CheckAccess(table, form.Columns)
	if err != nil {
		errorResponse(w, err)
		return
	}
	if len(form.Columns) > 0 {
		q = q.Select("id," + form.Columns)
	}

	result := new(listResult)
	err = q.Count(&result.Count).Error
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "table": table}).Error("Getting table records count")
		errorResponse(w, errTableNotFound.Errorf(table))
		return
	}

	rows, err := q.Order("id ASC").Offset(form.Offset).Limit(form.Limit).Rows()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "table": table}).Error("Getting rows from table")
		errorResponse(w, err)
		return
	}

	result.List, err = model.GetResult(rows)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, result)
}
