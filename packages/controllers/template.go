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
	"github.com/DayLightProject/go-daylight/packages/utils"
)



func Template(w http.ResponseWriter, r *http.Request) {
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
	sessWalletId := GetSessWalletId(sess)
	sessCitizenId := GetSessCitizenId(sess)
	sessStateId := GetSessInt64("state_id", sess)
	sessAccountId := GetSessInt64("account_id", sess)
	//sessAddress := GetSessString(sess, "address")
	log.Debug("sessWalletId %v / sessCitizenId %v", sessWalletId, sessCitizenId)

	r.ParseForm()
	page := r.FormValue("page")
	if len(page) == 0 {
		log.Error("%v", len(page) == 0)
		return
	}

	tpl, err := utils.CreateHtmlFromTemplate(page, sessCitizenId, sessAccountId, sessStateId)
	if err != nil {
		log.Error("%v", err)
		return
	}
	w.Write([]byte(tpl))
	return
}
