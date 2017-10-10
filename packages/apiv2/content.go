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

package apiv2

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/templatev2"
)

type contentResult struct {
	Tree string `json:"tree"`
}

func initVars(data *apiData) *map[string]string {
	vars := make(map[string]string)
	vars[`state`] = converter.Int64ToStr(data.state)
	vars[`wallet`] = converter.Int64ToStr(data.wallet)
	return &vars
}

func getPage(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var query string
	params := make(map[string]string)
	page := data.params[`name`].(string)
	params[`accept_lang`] = r.Header.Get(`Accept-Language`)
	query = `SELECT value FROM "` + converter.Int64ToStr(data.state) + `_pages" WHERE name = ?`
	pattern, err := model.Single(query, page).String()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting page")
		return err
	}
	ret := templatev2.Template2JSON(pattern, false, initVars(data))
	data.result = &contentResult{Tree: string(ret)}
	return nil
}

func getMenu(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var query string
	params := make(map[string]string)
	params[`accept_lang`] = r.Header.Get(`Accept-Language`)
	query = `SELECT value FROM "` + converter.Int64ToStr(data.state) + `_menu" WHERE name = ?`
	pattern, err := model.Single(query, data.params[`name`].(string)).String()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting menu")
		return errorAPI(w, err, http.StatusBadRequest)
	}
	ret := templatev2.Template2JSON(pattern, false, initVars(data))
	data.result = &contentResult{Tree: string(ret)}
	return nil
}

func jsonContent(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	ret := templatev2.Template2JSON(data.params[`template`].(string), false, initVars(data))
	data.result = &contentResult{Tree: string(ret)}
	return nil
}
