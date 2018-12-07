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
	"encoding/hex"
	"encoding/json"
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

func initVars(r *http.Request, data *apiData) *map[string]string {
	vars := make(map[string]string)
	for name := range r.Form {
		vars[name] = r.FormValue(name)
	}
	vars[`_full`] = `0`
	vars[`guest_key`] = consts.GuestKey
	if data.keyId != 0 {
		vars[`ecosystem_id`] = converter.Int64ToStr(data.ecosystemId)
		vars[`key_id`] = converter.Int64ToStr(data.keyId)
		vars[`isMobile`] = data.isMobile
		vars[`role_id`] = converter.Int64ToStr(data.roleId)
		vars[`ecosystem_name`] = data.ecosystemName
	} else {
		vars[`ecosystem_id`] = vars[`ecosystem`]
		if len(vars[`keyID`]) > 0 {
			vars[`key_id`] = vars[`keyID`]
		} else {
			vars[`key_id`] = `0`
		}
		if len(vars[`roleID`]) > 0 {
			vars[`role_id`] = vars[`roleID`]
		} else {
			vars[`role_id`] = `0`
		}
		if len(vars[`isMobile`]) == 0 {
			vars[`isMobile`] = `0`
		}
		if len(vars[`ecosystem_id`]) != 0 {
			ecosystems := model.Ecosystem{}
			if found, _ := ecosystems.Get(converter.StrToInt64(vars[`ecosystem_id`])); found {
				vars[`ecosystem_name`] = ecosystems.Name
			}
		}
	}
	if _, ok := vars[`lang`]; !ok {
		vars[`lang`] = r.Header.Get(`Accept-Language`)
	}

	return &vars
}

func parseEcosystem(in string) (string, string) {
	ecosystem, name := converter.ParseName(in)
	if ecosystem == 0 {
		return ``, name
	}
	return converter.Int64ToStr(ecosystem), name
}

func pageValue(w http.ResponseWriter, data *apiData, logger *log.Entry) (*model.Page, string, error) {
	var ecosystem string
	page := &model.Page{}
	name := data.params[`name`].(string)
	if strings.HasPrefix(name, `@`) {
		ecosystem, name = parseEcosystem(name)
		if len(name) == 0 {
			logger.WithFields(log.Fields{"type": consts.NotFound,
				"value": data.params[`name`].(string)}).Error("page not found")
			return nil, ``, errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
		}
	} else {
		ecosystem = getPrefix(data)
	}
	page.SetTablePrefix(ecosystem)
	found, err := page.Get(name)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting page")
		return nil, ``, errorAPI(w, `E_SERVER`, http.StatusInternalServerError)
	}
	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound}).Error("page not found")
		return nil, ``, errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
	}
	return page, ecosystem, nil
}

func getPage(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	page, prefix, err := pageValue(w, data, logger)
	if err != nil {
		return err
	}
	menu := &model.Menu{}
	menu.SetTablePrefix(prefix)
	_, err = menu.Get(getPrefix(data))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting page menu")
		return errorAPI(w, `E_SERVER`, http.StatusInternalServerError)
	}
	var wg sync.WaitGroup
	var timeout bool
	wg.Add(2)
	success := make(chan bool, 1)
	go func() {
		defer wg.Done()

		vars := initVars(r, data)
		(*vars)["app_id"] = converter.Int64ToStr(page.AppID)

		ret := template.Template2JSON(page.Value, &timeout, vars)
		if timeout {
			return
		}
		retmenu := template.Template2JSON(menu.Value, &timeout, vars)
		if timeout {
			return
		}
		data.result = &contentResult{Tree: ret, Menu: page.Menu, MenuTree: retmenu, NodesCount: page.ValidateCount}
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
		log.WithFields(log.Fields{"type": consts.InvalidObject}).Error(page.Name + " is a heavy page")
		return errorAPI(w, `E_HEAVYPAGE`, http.StatusInternalServerError)
	}
	return nil
}

func getPageHash(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	if ecosystem := r.FormValue(`ecosystem`); len(ecosystem) > 0 &&
		!strings.HasPrefix(data.params[`name`].(string), `@`) {
		data.params[`name`] = `@` + ecosystem + data.params[`name`].(string)
	}
	err = getPage(w, r, data, logger)
	if err == nil {
		var out, ret []byte
		out, err = json.Marshal(data.result.(*contentResult))
		if err != nil {
			log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("getting string for hash")
			return errorAPI(w, `E_SERVER`, http.StatusInternalServerError)
		}
		ret, err = crypto.Hash(out)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("calculating hash of the page")
			return errorAPI(w, `E_SERVER`, http.StatusInternalServerError)
		}
		data.result = &hashResult{Hash: hex.EncodeToString(ret)}
	}
	return
}

func getMenu(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var ecosystem string
	menu := &model.Menu{}
	name := data.params[`name`].(string)
	if strings.HasPrefix(name, `@`) {
		ecosystem, name = parseEcosystem(name)
		if len(name) == 0 {
			logger.WithFields(log.Fields{"type": consts.NotFound,
				"value": data.params[`name`].(string)}).Error("page not found")
			return errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
		}
	} else {
		ecosystem = getPrefix(data)
	}

	menu.SetTablePrefix(ecosystem)
	found, err := menu.Get(name)

	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting menu")
		return errorAPI(w, err, http.StatusBadRequest)
	}
	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound}).Error("menu not found")
		return errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
	}
	var timeout bool
	ret := template.Template2JSON(menu.Value, &timeout, initVars(r, data))
	data.result = &contentResult{Tree: ret, Title: menu.Title}
	return nil
}

func jsonContent(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var timeout bool
	vars := initVars(r, data)
	if data.params[`source`].(string) == strOne || data.params[`source`].(string) == strTrue {
		(*vars)["_full"] = strOne
	}
	ret := template.Template2JSON(data.params[`template`].(string), &timeout, vars)
	data.result = &contentResult{Tree: ret}
	return nil
}

func getSource(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	page, _, err := pageValue(w, data, logger)
	if err != nil {
		return err
	}
	var timeout bool
	vars := initVars(r, data)
	(*vars)["_full"] = strOne
	ret := template.Template2JSON(page.Value, &timeout, vars)
	data.result = &contentResult{Tree: ret}
	return nil
}
