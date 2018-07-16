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
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/template"

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
	strOne = `1`
)

func initVars(r *http.Request) *map[string]string {
	client := getClient(r)

	r.ParseForm()
	vars := make(map[string]string)
	for name := range r.Form {
		vars[name] = r.FormValue(name)
	}
	vars[`_full`] = `0`
	if client.KeyID != 0 {
		vars[`ecosystem_id`] = converter.Int64ToStr(client.EcosystemID)
		vars[`key_id`] = converter.Int64ToStr(client.KeyID)
		vars[`isMobile`] = client.IsMobile
		vars[`role_id`] = converter.Int64ToStr(client.RoleID)
		vars[`ecosystem_name`] = client.EcosystemName
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

func pageValue(r *http.Request) (*model.Page, error) {
	params := mux.Vars(r)
	logger := getLogger(r)
	client := getClient(r)

	page := &model.Page{}
	page.SetTablePrefix(client.Prefix())
	if found, err := page.Get(params[keyName]); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting page")
		return nil, errServer
	} else if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound}).Error("page not found")
		return nil, errNotFound
	}
	return page, nil
}

func getPage(r *http.Request) (result *contentResult, err error) {
	page, err := pageValue(r)
	if err != nil {
		return nil, err
	}

	client := getClient(r)
	logger := getLogger(r)

	menu := model.Menu{}
	menu.SetTablePrefix(client.Prefix())
	if _, err = menu.Get(page.Menu); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting single from DB")
		return nil, errServer
	}

	var (
		wg      sync.WaitGroup
		timeout bool
	)

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
	result, err := getPage(r)
	if err != nil {
		errorResponse(w, err)
		return
	}

	logger := getLogger(r)

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

	menu := &model.Menu{}
	menu.SetTablePrefix(client.Prefix())
	if found, err := menu.Get(params[keyName]); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting menu")
		errorResponse(w, newError(err, http.StatusBadRequest))
		return
	} else if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound}).Error("menu not found")
		errorResponse(w, errNotFound)
		return
	}

	var timeout bool
	ret := template.Template2JSON(menu.Value, &timeout, initVars(r))

	jsonResponse(w, &contentResult{
		Tree:  ret,
		Title: menu.Title,
	})
}

type jsonContentForm struct {
	form
	Template string `schema:"template"`
	Source   bool   `schema:"source"`
}

func (f *jsonContentForm) Validate(r *http.Request) error {
	if len(f.Template) == 0 {
		return newError(fmt.Errorf("Empty template"), http.StatusBadRequest)
	}
	return nil
}

func jsonContentHandler(w http.ResponseWriter, r *http.Request) {
	form := &jsonContentForm{}
	if err := parseForm(r, form); err != nil {
		errorResponse(w, err)
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
	page, err := pageValue(r)
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
