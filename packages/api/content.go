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
	strTrue = `true`
	strOne  = `1`
)

func initVars(r *http.Request) *map[string]string {
	client := getClient(r)

	r.ParseForm()
	vars := make(map[string]string)
	for name := range r.Form {
		vars[name] = r.FormValue(name)
	}
	vars[`_full`] = `0`
	vars[`ecosystem_id`] = converter.Int64ToStr(client.EcosystemID)
	vars[`key_id`] = converter.Int64ToStr(client.KeyID)
	vars[`isMobile`] = client.IsMobile
	vars[`role_id`] = converter.Int64ToStr(client.RoleID)
	vars[`ecosystem_name`] = client.EcosystemName

	if _, ok := vars[`lang`]; !ok {
		vars[`lang`] = r.Header.Get(`Accept-Language`)
	}

	return &vars
}

func pageValue(w http.ResponseWriter, r *http.Request) (*model.Page, bool) {
	params := mux.Vars(r)
	logger := getLogger(r)
	client := getClient(r)

	page := &model.Page{}
	page.SetTablePrefix(client.Prefix())
	if found, err := page.Get(params[keyName]); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting page")
		errorResponse(w, errServer, http.StatusInternalServerError)
		return nil, false
	} else if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound}).Error("page not found")
		errorResponse(w, errNotFound, http.StatusNotFound)
		return nil, false
	}
	return page, true
}

func getPage(w http.ResponseWriter, r *http.Request) (result *contentResult, ok bool) {
	page, ok := pageValue(w, r)
	if !ok {
		return
	}

	client := getClient(r)
	logger := getLogger(r)

	// TODO: перенести в модели
	menu, err := model.Single(`SELECT value FROM "`+client.Prefix()+`_menu" WHERE name = ?`,
		page.Menu).String()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting single from DB")
		errorResponse(w, errServer, http.StatusInternalServerError)
		return
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
		retmenu := template.Template2JSON(menu, &timeout, vars)
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
		errorResponse(w, errHeavyPage, http.StatusInternalServerError)
		return
	}

	return result, true
}

func getPageHandler(w http.ResponseWriter, r *http.Request) {
	result, ok := getPage(w, r)
	if !ok {
		return
	}

	jsonResponse(w, result)
}

func getPageHashHandler(w http.ResponseWriter, r *http.Request) {
	result, ok := getPage(w, r)
	if !ok {
		return
	}

	logger := getLogger(r)

	out, err := json.Marshal(result)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("getting string for hash")
		errorResponse(w, errServer, http.StatusInternalServerError)
		return
	}
	ret, err := crypto.Hash(out)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("calculating hash of the page")
		errorResponse(w, errServer, http.StatusInternalServerError)
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
		errorResponse(w, err, http.StatusBadRequest)
	} else if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound}).Error("menu not found")
		errorResponse(w, errNotFound, http.StatusNotFound)
	}

	var timeout bool
	ret := template.Template2JSON(menu.Value, &timeout, initVars(r))

	jsonResponse(w, &contentResult{
		Tree:  ret,
		Title: menu.Title,
	})
}

type jsonContentForm struct {
	Form
	Template string `schema:"template"`
	Source   string `schema:"source"`
}

func (f *jsonContentForm) Validate(w http.ResponseWriter, r *http.Request) bool {
	if len(f.Template) == 0 {
		errorResponse(w, fmt.Errorf("Empty template"), http.StatusBadRequest)
		return false
	}
	return true
}

func jsonContentHandler(w http.ResponseWriter, r *http.Request) {
	form := &jsonContentForm{}
	if ok := ParseForm(w, r, form); !ok {
		return
	}

	var timeout bool
	vars := initVars(r)

	// TODO: bool
	if form.Source == strOne || form.Source == strTrue {
		(*vars)["_full"] = strOne
	}

	ret := template.Template2JSON(form.Template, &timeout, vars)
	jsonResponse(w, &contentResult{Tree: ret})
}

func getSourceHandler(w http.ResponseWriter, r *http.Request) {
	page, ok := pageValue(w, r)
	if !ok {
		return
	}

	var timeout bool
	vars := initVars(r)
	(*vars)["_full"] = strOne
	ret := template.Template2JSON(page.Value, &timeout, vars)

	jsonResponse(w, &contentResult{Tree: ret})
}
