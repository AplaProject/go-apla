// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package api

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/template"

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

func parseEcosystem(in string) (id, name string) {
	re := regexp.MustCompile(`(?is)^@(\d+)(\D.*)$`)
	ret := re.FindStringSubmatch(in)
	if len(ret) == 3 {
		id = ret[1]
		name = ret[2]
	}
	return
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
	_, err = menu.Get(page.Menu)
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
	menu := &model.Menu{}
	menu.SetTablePrefix(getPrefix(data))
	found, err := menu.Get(data.params[`name`].(string))

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
