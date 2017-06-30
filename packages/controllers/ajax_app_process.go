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

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
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
		table  string
	)
	name := c.r.FormValue("name")
	block := converter.StrToInt64(c.r.FormValue("block"))
	done := converter.StrToInt(c.r.FormValue("done"))
	if block == 0 {
		result.Error = `wrong block id`
		return result
	}
	if strings.HasPrefix(name, `global`) {
		table = `global_apps`
	} else {
		table = fmt.Sprintf(`"%d_apps"`, c.SessStateID)
	}
	cur, err := c.OneRow(`select * from `+table+` where name=?`, name).String()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if len(cur) > 0 {
		err = c.ExecSQL(fmt.Sprintf(`update %s set done=?, blocks=concat(blocks, ',%d') where name=?`, table, block),
			done, name)
	} else {
		err = c.ExecSQL(fmt.Sprintf(`insert into %s (name,done,blocks) values(?,?,'%d')`, table, block),
			name, done)
	}
	if err != nil {
		result.Error = err.Error()
		return result
	}

	return result
}
