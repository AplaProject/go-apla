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
	"fmt"
	"strings"

	"strconv"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
)

const aAppProcess = `ajax_app_process`

// AppProcess is a structure of the ajax_app_process ajax request
type AppProcess struct {
	Error string `json:"error"`
}

func init() {
	newPage(aAppProcess, `json`)
}

// AjaxAppProcess is a controller of ajax_app_process request
func (c *Controller) AjaxAppProcess() interface{} {
	var (
		result AppProcess
	)
	name := c.r.FormValue("name")
	block := converter.StrToInt64(c.r.FormValue("block"))
	done := converter.StrToInt(c.r.FormValue("done"))

	if block == 0 {
		result.Error = `wrong block id`
		return result
	}

	app := &model.App{Name: name, Done: int32(done)}
	if strings.HasPrefix(name, `global`) {
		app.SetTablePrefix("global")
	} else {
		app.SetTablePrefix(strconv.FormatInt(c.SessStateID, 10))
	}
	exist, err := app.IsExists(app.Name)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if exist {
		app.Blocks += fmt.Sprintf(",%d", block)
		err = app.Save()
	} else {
		app.Blocks = fmt.Sprintf("%d", block)
		err = app.Create()
	}
	if err != nil {
		result.Error = err.Error()
		return result
	}

	return result
}
