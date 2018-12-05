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

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/language"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

const defaultSectionsLimit = 100

type sectionsForm struct {
	paginatorForm
	Lang string `schema:"lang"`
}

func (f *sectionsForm) Validate(r *http.Request) error {
	if err := f.paginatorForm.Validate(r); err != nil {
		return err
	}

	if len(f.Lang) == 0 {
		f.Lang = r.Header.Get("Accept-Language")
	}

	return nil
}

func getSectionsHandler(w http.ResponseWriter, r *http.Request) {
	form := &sectionsForm{}
	form.defaultLimit = defaultSectionsLimit
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	client := getClient(r)
	logger := getLogger(r)

	table := "1_sections"
	q := model.GetDB(nil).Table(table).Where("ecosystem = ? AND status > 0", client.EcosystemID).Order("id ASC")

	result := new(listResult)
	err := q.Count(&result.Count).Error
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "table": table}).Error("Getting table records count")
		errorResponse(w, errTableNotFound.Errorf(table))
		return
	}

	rows, err := q.Offset(form.Offset).Limit(form.Limit).Rows()
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

	var sections []map[string]string
	for _, item := range result.List {
		var roles []int64
		if err := json.Unmarshal([]byte(item["roles_access"]), &roles); err != nil {
			errorResponse(w, err)
			return
		}
		if len(roles) > 0 {
			var added bool
			for _, v := range roles {
				if v == client.RoleID {
					added = true
					break
				}
			}
			if !added {
				continue
			}
		}

		if item["status"] == "2" {
			roles := &model.Role{}
			roles.SetTablePrefix("1")
			role, err := roles.Get(nil, client.RoleID)

			if err != nil {
				logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "table": table}).Error("Getting role by id")
				errorResponse(w, err)
				return
			}
			if role == true {
				item["default_page"] = roles.DefaultPage
			}
		}
		
		item["title"] = language.LangMacro(item["title"], int(client.EcosystemID), form.Lang)
		sections = append(sections, item)
	}
	result.List = sections

	jsonResponse(w, result)
}
