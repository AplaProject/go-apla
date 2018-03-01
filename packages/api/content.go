// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package api

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
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
	Menu     string          `json:"menu,omitempty"`
	MenuTree json.RawMessage `json:"menutree,omitempty"`
	Title    string          `json:"title,omitempty"`
	Tree     json.RawMessage `json:"tree"`
}

type hashResult struct {
	Hash string `json:"hash"`
}

func initVars(r *http.Request, data *apiData) *map[string]string {
	vars := make(map[string]string)
	for name := range r.Form {
		vars[name] = r.FormValue(name)
	}
	vars[`_full`] = `0`
	vars[`ecosystem_id`] = converter.Int64ToStr(data.ecosystemId)
	vars[`key_id`] = converter.Int64ToStr(data.keyId)

	if _, ok := vars[`lang`]; !ok {
		vars[`lang`] = r.Header.Get(`Accept-Language`)
	}

	return &vars
}

func pageValue(w http.ResponseWriter, data *apiData, logger *log.Entry) (*model.Page, error) {
	page := &model.Page{}
	page.SetTablePrefix(getPrefix(data))
	found, err := page.Get(data.params[`name`].(string))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting page")
		return nil, errorAPI(w, `E_SERVER`, http.StatusInternalServerError)
	}
	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound}).Error("page not found")
		return nil, errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
	}
	return page, nil
}

func getPage(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {

	page, err := pageValue(w, data, logger)
	if err != nil {
		return err
	}
	menu, err := model.Single(`SELECT value FROM "`+getPrefix(data)+`_menu" WHERE name = ?`,
		page.Menu).String()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting single from DB")
		return errorAPI(w, `E_SERVER`, http.StatusInternalServerError)
	}
	var wg sync.WaitGroup
	var timeout bool
	wg.Add(2)
	success := make(chan bool, 1)
	go func() {
		defer wg.Done()

		ret := template.Template2JSON(page.Value, &timeout, initVars(r, data))
		if timeout {
			return
		}
		retmenu := template.Template2JSON(menu, &timeout, initVars(r, data))
		if timeout {
			return
		}
		data.result = &contentResult{Tree: ret, Menu: page.Menu, MenuTree: retmenu}
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
	ret := template.Template2JSON(data.params[`template`].(string), &timeout, initVars(r, data))
	data.result = &contentResult{Tree: ret}
	return nil
}

func getSource(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	page, err := pageValue(w, data, logger)
	if err != nil {
		return err
	}
	var timeout bool
	vars := initVars(r, data)
	(*vars)["_full"] = "1"
	ret := template.Template2JSON(page.Value, &timeout, vars)
	data.result = &contentResult{Tree: ret}
	return nil
}
