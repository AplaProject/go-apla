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

	"github.com/EGaaS/go-egaas-mvp/packages/static"
	"github.com/EGaaS/go-egaas-mvp/packages/textproc"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

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
	params[`state_id`] = utils.Int64ToStr(GetSessInt64("state_id", sess))
	params[`wallet_id`] = utils.Int64ToStr(GetSessWalletId(sess))
	params[`citizen_id`] = utils.Int64ToStr(GetSessCitizenId(sess))

	var out string
	data, err := static.Asset("static/" + page + ".tpl")
	if err != nil {
		out = err.Error()
	}

	if len(data) > 0 {
		out, _ = utils.ProceedTemplate(`app_template`, &utils.PageTpl{Page: page,
			Template: textproc.Process(string(data), &params),
			Data: &utils.CommonPage{
				WalletId:     GetSessWalletId(sess),
				CitizenId:    GetSessCitizenId(sess),
				StateId:      GetSessInt64("state_id", sess),
				CountSignArr: []int{0}}})
	}

	w.Write([]byte(out))
	return
}
