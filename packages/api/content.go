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
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/template"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type contentResult struct {
	Menu       string          `json:"menu,omitempty"`
	MenuTree   json.RawMessage `json:"menutree,omitempty"`
	Title      string          `json:"title,omitempty"`
	Tree       json.RawMessage `json:"tree"`
	NodesCount int64           `json:"nodesCount,omitempty"`
}

type hashResult struct {
	Hash string `json:"hash"`
}

const (
	strTrue = `true`
	strOne  = `1`
)

var errEmptyTemplate = errors.New("Empty template")

func initVars(r *http.Request) *map[string]string {
	client := getClient(r)
	r.ParseMultipartForm(multipartBuf)

	vars := make(map[string]string)
	for name := range r.Form {
		vars[name] = r.FormValue(name)
	}
	vars["_full"] = "0"
	vars["guest_key"] = consts.GuestKey
	vars["guest_account"] = consts.GuestAddress
	if client.KeyID != 0 {
		vars["ecosystem_id"] = converter.Int64ToStr(client.EcosystemID)
		vars["key_id"] = converter.Int64ToStr(client.KeyID)
		vars["account_id"] = client.AccountID
		vars["isMobile"] = isMobileValue(client.IsMobile)
		vars["role_id"] = converter.Int64ToStr(client.RoleID)
		vars["ecosystem_name"] = client.EcosystemName
	} else {
		vars["ecosystem_id"] = vars["ecosystem"]
		delete(vars, "ecosystem")
		if len(vars["keyID"]) > 0 {
			vars["key_id"] = vars["keyID"]
			vars["account_id"] = converter.AddressToString(converter.StrToInt64(vars["keyID"]))
		} else {
			vars["key_id"] = "0"
			vars["account_id"] = ""
		}
		if len(vars["roleID"]) > 0 {
			vars["role_id"] = vars["roleID"]
		} else {
			vars["role_id"] = "0"
		}
		if len(vars["isMobile"]) == 0 {
			vars["isMobile"] = "0"
		}
		if len(vars["ecosystem_id"]) != 0 {
			ecosystems := model.Ecosystem{}
			if found, _ := ecosystems.Get(converter.StrToInt64(vars["ecosystem_id"])); found {
				vars["ecosystem_name"] = ecosystems.Name
			}
		}
	}
	if _, ok := vars["lang"]; !ok {
		vars["lang"] = r.Header.Get("Accept-Language")
	}

	return &vars
}

func isMobileValue(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func parseEcosystem(in string) (string, string) {
	ecosystem, name := converter.ParseName(in)
	if ecosystem == 0 {
		return ``, name
	}
	return converter.Int64ToStr(ecosystem), name
}

func pageValue(r *http.Request) (*model.Page, string, error) {
	params := mux.Vars(r)
	logger := getLogger(r)
	client := getClient(r)

	var ecosystem string
	page := &model.Page{}
	name := params["name"]
	if strings.HasPrefix(name, `@`) {
		ecosystem, name = parseEcosystem(name)
		if len(name) == 0 {
			logger.WithFields(log.Fields{
				"type":  consts.NotFound,
				"value": params["name"],
			}).Error("page not found")
			return nil, ``, errNotFound
		}
	} else {
		ecosystem = client.Prefix()
	}
	page.SetTablePrefix(ecosystem)
	found, err := page.Get(name)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting page")
		return nil, ``, err
	}
	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound}).Error("page not found")
		return nil, ``, errNotFound
	}
	return page, ecosystem, nil
}

func getPage(r *http.Request) (result *contentResult, err error) {
	page, _, err := pageValue(r)
	if err != nil {
		return nil, err
	}

	logger := getLogger(r)

	client := getClient(r)
	menu := &model.Menu{}
	menu.SetTablePrefix(client.Prefix())
	_, err = menu.Get(page.Menu)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting page menu")
		return nil, errServer
	}
	var wg sync.WaitGroup
	var timeout bool
	wg.Add(2)
	success := make(chan bool, 1)
	go func() {
		defer wg.Done()

		vars := initVars(r)
		(*vars)["app_id"] = converter.Int64ToStr(page.AppID)

		ret := template.Template2JSON(page.Value, &timeout, vars)
		if timeout {
			return
		}
		retmenu := template.Template2JSON(menu.Value, &timeout, vars)
		if timeout {
			return
		}
		result = &contentResult{
			Tree:       ret,
			Menu:       page.Menu,
			MenuTree:   retmenu,
			NodesCount: page.ValidateCount,
		}
		success <- true
	}()
	go func() {
		defer wg.Done()
		if conf.Config.MaxPageGenerationTime == 0 {
			return
		}
		select {
		case <-time.After(time.Duration(conf.Config.MaxPageGenerationTime) * time.Millisecond):
			timeout = true
		case <-success:
		}
	}()
	wg.Wait()
	close(success)

	if timeout {
		logger.WithFields(log.Fields{"type": consts.InvalidObject}).Error(page.Name + " is a heavy page")
		return nil, errHeavyPage
	}

	return result, nil
}

func getPageHandler(w http.ResponseWriter, r *http.Request) {
	result, err := getPage(r)
	if err != nil {
		errorResponse(w, err)
		return
	}

	jsonResponse(w, result)
}

func getPageHashHandler(w http.ResponseWriter, r *http.Request) {
	logger := getLogger(r)
	params := mux.Vars(r)

	if ecosystem := r.FormValue("ecosystem"); len(ecosystem) > 0 &&
		!strings.HasPrefix(params["name"], "@") {
		params["name"] = "@" + ecosystem + params["name"]
	}
	result, err := getPage(r)
	if err != nil {
		errorResponse(w, err)
		return
	}

	out, err := json.Marshal(result)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("getting string for hash")
		errorResponse(w, errServer)
		return
	}
	ret, err := crypto.Hash(out)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("calculating hash of the page")
		errorResponse(w, errServer)
		return
	}

	jsonResponse(w, &hashResult{Hash: hex.EncodeToString(ret)})
}

func getMenuHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	client := getClient(r)
	logger := getLogger(r)

	var ecosystem string
	menu := &model.Menu{}
	name := params["name"]
	if strings.HasPrefix(name, `@`) {
		ecosystem, name = parseEcosystem(name)
		if len(name) == 0 {
			logger.WithFields(log.Fields{
				"type":  consts.NotFound,
				"value": params["name"],
			}).Error("page not found")
			errorResponse(w, errNotFound)
			return
		}
	} else {
		ecosystem = client.Prefix()
	}

	menu.SetTablePrefix(ecosystem)
	found, err := menu.Get(name)

	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting menu")
		errorResponse(w, err)
		return
	}
	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound}).Error("menu not found")
		errorResponse(w, errNotFound)
		return
	}
	var timeout bool
	ret := template.Template2JSON(menu.Value, &timeout, initVars(r))
	jsonResponse(w, &contentResult{Tree: ret, Title: menu.Title})
}

type jsonContentForm struct {
	Template string `schema:"template"`
	Source   bool   `schema:"source"`
}

func (f *jsonContentForm) Validate(r *http.Request) error {
	if len(f.Template) == 0 {
		return errEmptyTemplate
	}
	return nil
}

func jsonContentHandler(w http.ResponseWriter, r *http.Request) {
	form := &jsonContentForm{}
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err, http.StatusBadRequest)
		return
	}

	var timeout bool
	vars := initVars(r)

	if form.Source {
		(*vars)["_full"] = strOne
	}

	ret := template.Template2JSON(form.Template, &timeout, vars)
	jsonResponse(w, &contentResult{Tree: ret})
}

func getSourceHandler(w http.ResponseWriter, r *http.Request) {
	page, _, err := pageValue(r)
	if err != nil {
		errorResponse(w, err)
		return
	}
	var timeout bool
	vars := initVars(r)
	(*vars)["_full"] = strOne
	ret := template.Template2JSON(page.Value, &timeout, vars)

	jsonResponse(w, &contentResult{Tree: ret})
}

func getPageValidatorsCountHandler(w http.ResponseWriter, r *http.Request) {
	page, _, err := pageValue(r)
	if err != nil {
		errorResponse(w, err)
		return
	}

	res := map[string]int64{"validate_count": page.ValidateCount}
	jsonResponse(w, &res)
}
