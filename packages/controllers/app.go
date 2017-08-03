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
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/static"
	"github.com/EGaaS/go-egaas-mvp/packages/template"
	"github.com/EGaaS/go-egaas-mvp/packages/textproc"
)

type appData struct {
	template.CommonPage
	Done    bool
	Proceed int
	Blocks  []string
}

// App is a controller for application install template page
func App(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Error("Content Recovered", fmt.Sprintf("%s: %s", e, debug.Stack()))
			fmt.Println("Content Recovered", fmt.Sprintf("%s: %s", e, debug.Stack()))
		}
	}()
	var err error

	w.Header().Set("Content-type", "text/html")

	sess, err := globalSessions.SessionStart(w, r)
	if err != nil {
		log.Error("%v", err)
	}
	defer sess.SessionRelease(w)
	r.ParseForm()
	page := r.FormValue("page")
	params := make(map[string]string)
	if len(page) == 0 {
		log.Error("%v", len(page) == 0)
		return
	}
	for name := range r.Form {
		params[name] = r.FormValue(name)
	}

	params[`name`] = page
	params[`state_id`] = converter.Int64ToStr(GetSessInt64("state_id", sess))
	params[`wallet_id`] = converter.Int64ToStr(GetSessWalletID(sess))
	params[`citizen_id`] = converter.Int64ToStr(GetSessCitizenID(sess))

	var (
		out  string
		data []byte
	)

	if len(params[`file`]) == 0 {
		data, err = static.Asset("static/" + page + ".tpl")
		if err != nil {
			out = err.Error()
		}
	} else {
		data = []byte(params[`file`])
	}
	if len(data) > 0 {
		var table string
		if strings.HasPrefix(page, `global`) {
			table = `global`
		} else {
			table = fmt.Sprintf(`"%d"`, GetSessInt64("state_id", sess))
		}
		app := &model.App{}
		app.SetTablePrefix(table)
		err := app.Get(page)
		if err != nil {
			out = err.Error()
		} else {
			var done bool
			var blocks []string
			done = app.Done == 1
			blocks = strings.Split(app.Blocks, `,`)
			out, _ = template.ProceedTemplate(`app_template`, &template.PageTpl{Page: page,
				Template: textproc.Process(string(data), &params), Unique: ``,
				Data: &appData{
					CommonPage: template.CommonPage{WalletId: GetSessWalletID(sess),
						CitizenId: GetSessCitizenID(sess),
						StateId:   GetSessInt64("state_id", sess),
					},
					Blocks:  blocks,
					Proceed: len(blocks),
					Done:    done,
				}})
		}
	}

	w.Write([]byte(out))
	return
}
