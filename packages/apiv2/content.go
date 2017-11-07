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

	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/templatev2"
)

type contentResult struct {
	Menu     string `json:"menu,omitempty"`
	MenuTree string `json:"menutree,omitempty"`
	Title    string `json:"title,omitempty"`
	Tree     string `json:"tree"`
}

func initVars(r *http.Request, data *apiData) *map[string]string {
	vars := make(map[string]string)
	for name := range r.Form {
		vars[name] = r.FormValue(name)
	}
	vars[`ecosystem_id`] = converter.Int64ToStr(data.ecosystemId)
	vars[`key_id`] = converter.Int64ToStr(data.keyId)
	vars[`accept_lang`] = r.Header.Get(`Accept-Language`)
	return &vars
}

func getPage(w http.ResponseWriter, r *http.Request, data *apiData) error {

	page := &model.Page{}
	page.SetTablePrefix(converter.Int64ToStr(data.ecosystemId))
	found, err := page.Get(data.params[`name`].(string))
	if err != nil {
		return err
	}
	if !found {
		return errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
	}

	ret := templatev2.Template2JSON(page.Value, false, initVars(r, data))

	menu, err := model.Single(`SELECT value FROM "`+converter.Int64ToStr(data.ecosystemId)+
		`_menu" WHERE name = ?`, page.Menu).String()
	retmenu := templatev2.Template2JSON(menu, false, initVars(r, data))

	data.result = &contentResult{Tree: string(ret), Menu: page.Menu, MenuTree: string(retmenu)}
	return nil
}

func getMenu(w http.ResponseWriter, r *http.Request, data *apiData) error {
	menu := &model.Menu{}
	menu.SetTablePrefix(converter.Int64ToStr(data.ecosystemId))
	found, err := menu.Get(data.params[`name`].(string))

	if err != nil {
		return errorAPI(w, err, http.StatusBadRequest)
	}
	if !found {
		return errorAPI(w, `E_NOTFOUND`, http.StatusNotFound)
	}

	ret := templatev2.Template2JSON(menu.Value, false, initVars(r, data))
	data.result = &contentResult{Tree: string(ret), Title: menu.Title}
	return nil
}

func jsonContent(w http.ResponseWriter, r *http.Request, data *apiData) error {
	ret := templatev2.Template2JSON(data.params[`template`].(string), false, initVars(r, data))
	data.result = &contentResult{Tree: string(ret)}
	return nil
}
