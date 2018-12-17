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

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func (m Mode) getEcosystemParamHandler(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)

	form := &ecosystemForm{
		Validator: m.EcosysIDValidator,
	}

	if err := parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	params := mux.Vars(r)

	sp := &model.StateParameter{}
	sp.SetTablePrefix(form.EcosystemPrefix)
	name := params["name"]

	if found, err := sp.Get(nil, name); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting state parameter by name")
		errorResponse(w, err)
		return
	} else if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound, "key": name}).Error("state parameter not found")
		errorResponse(w, errParamNotFound.Errorf(name))
		return
	}

	jsonResponse(w, &paramResult{
		ID:         converter.Int64ToStr(sp.ID),
		Name:       sp.Name,
		Value:      sp.Value,
		Conditions: sp.Conditions,
	})
}

func getEcosystemNameHandler(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)

	ecosystemID := converter.StrToInt64(r.FormValue("id"))
	ecosystems := model.Ecosystem{}
	found, err := ecosystems.Get(ecosystemID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on getting ecosystem name")
		errorResponse(w, err)
		return
	}
	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound, "ecosystem_id": ecosystemID}).Error("ecosystem by id not found")
		errorResponse(w, errParamNotFound.Errorf("name"))
		return
	}

	jsonResponse(w, &struct {
		EcosystemName string `json:"ecosystem_name"`
	}{
		EcosystemName: ecosystems.Name,
	})
}
