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
	"net/http"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/language"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/template"
	"github.com/EGaaS/go-egaas-mvp/packages/textproc"
)

type contentResult struct {
	HTML string `json:"html"`
}

func contentPage(w http.ResponseWriter, r *http.Request, data *apiData) error {

	params := make(map[string]string)
	for name := range r.Form {
		params[name] = r.FormValue(name)
	}
	page := data.params[`page`].(string)
	if page == `body` {
		params[`autobody`] = r.FormValue("body")
	}
	params[`global`] = converter.Int64ToStr(data.params[`global`].(int64))
	params[`accept_lang`] = r.Header.Get(`Accept-Language`)
	tpl, err := template.CreateHTMLFromTemplate(page, data.sess.Get(`citizen`).(int64),
		data.sess.Get(`state`).(int64), &params)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}
	data.result = &contentResult{HTML: string(tpl)}
	return nil
}

func contentMenu(w http.ResponseWriter, r *http.Request, data *apiData) error {

	prefix := getPrefix(data)
	menu, err := model.Single(`SELECT value FROM "`+prefix+`_menu" WHERE name = ?`, data.params[`name`].(string)).String()
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}
	params := make(map[string]string)
	params[`state_id`] = converter.Int64ToStr(data.sess.Get(`state`).(int64))
	params[`global`] = converter.Int64ToStr(data.params[`global`].(int64))
	params[`accept_lang`] = r.Header.Get(`Accept-Language`)
	if len(menu) > 0 {
		menu = language.LangMacro(textproc.Process(menu, &params), int(data.sess.Get(`state`).(int64)), params[`accept_lang`]) +
			`<!--#` + data.params[`name`].(string) + `#-->`
	}
	data.result = &contentResult{HTML: menu}
	return nil
}
