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

package controllers

import (
	"strconv"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/language"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/textproc"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// AjaxGetMenuHtml is a controller for AjaxGetMenuHtml
func (c *Controller) AjaxGetMenuHtml() (string, error) {

	pageName := c.r.FormValue("page")

	global := c.r.FormValue("global")
	prefix := "global"
	if global == "" || global == "0" {
		prefix = c.StateIDStr
	}
	var err error
	page := &model.Page{}
	menu := &model.Menu{}
	if len(prefix) > 0 {
		page.SetTablePrefix(prefix)
		err = page.Get(pageName)
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		menu.SetTablePrefix(prefix)
		err = menu.Get(page.Menu)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}
	params := make(map[string]string)
	params[`state_id`] = c.StateIDStr
	params[`accept_lang`] = c.r.Header.Get(`Accept-Language`)
	if len(menu.Value) > 0 {
		stateID, err := strconv.Atoi(c.StateIDStr)
		if err != nil {
			logger.LogInfo(consts.StrtoInt64Error, c.StateIDStr)
		}
		menu.Value = language.LangMacro(textproc.Process(menu.Value, &params), stateID, params[`accept_lang`]) +
			`<!--#` + page.Menu + `#-->`
	}
	return menu.Value, nil

}
