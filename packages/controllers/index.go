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
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/static"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

type index struct {
	DbOk        bool
	Lang        map[string]string
	Key         string
	PKey        string
	State       string
	SetLang     string
	Accounts    string
	Thrust      bool
	IOS         bool
	Android     bool
	Mobile      bool
	ShowIOSMenu bool
	Version     string
	Langs       string
	LogoExt     string
}

// Index is a control for index page
func Index(w http.ResponseWriter, r *http.Request) {

	accounts, _ := ioutil.ReadFile(filepath.Join(*utils.Dir, `accounts.txt`))

	r.ParseForm()
	if _, ok := r.Form[``]; ok {
		expiration := time.Now().Add(32 * 24 * time.Hour)
		http.SetCookie(w, &http.Cookie{Name: "ref", Value: r.Form.Get(``), Expires: expiration})
	}

	params := make(map[string]interface{})
	if len(r.PostFormValue("parameters")) > 0 {
		err := json.Unmarshal([]byte(r.PostFormValue("parameters")), &params)
		if err != nil {
			log.Error("%v", err)
		}
		log.Debug("params=%", params)
	}
	parameters := make(map[string]string)
	for k, v := range params {
		parameters[k] = utils.InterfaceToStr(v)
	}

	lang := GetLang(w, r, parameters)

	sess, err := globalSessions.SessionStart(w, r)
	if err != nil {
		log.Error("%v", err)
	}
	defer sess.SessionRelease(w)

	sessCitizenID := GetSessCitizenID(sess)
	sessWalletID := GetSessWalletID(sess)
	//	var key string

	showIOSMenu := true
	// Когда меню не выдаем
	// When we don't give the menu
	if utils.DB == nil || utils.DB.DB == nil {
		showIOSMenu = false
	}

	if sessCitizenID == 0 && sessWalletID == 0 {
		showIOSMenu = false
	}

	if showIOSMenu && utils.DB != nil && utils.DB.DB != nil {
		blockData, err := utils.DB.GetInfoBlock()
		if err != nil {
			log.Error("%v", err)
		}
		wTime := int64(12)
		wTimeReady := int64(2)
		log.Debug("wTime: %v / utils.Time(): %v / blockData[time]: %v", wTime, utils.Time(), utils.StrToInt64(blockData["time"]))
		// если время менее 12 часов от текущего, то выдаем не подвержденные, а просто те, что есть в блокчейне
		// if time differs less than for 12 hours from current time, give not affected but those which are in blockchain
		if utils.Time()-utils.StrToInt64(blockData["time"]) < 3600*wTime {
			lastBlockData, err := utils.DB.GetLastBlockData()
			if err != nil {
				log.Error("%v", err)
			}
			log.Debug("lastBlockData[lastBlockTime]: %v", lastBlockData["lastBlockTime"])
			log.Debug("time.Now().Unix(): %v", utils.Time())
			if utils.Time()-lastBlockData["lastBlockTime"] >= 3600*wTimeReady {
				showIOSMenu = false
			}
		} else {
			showIOSMenu = false
		}
	}
	if showIOSMenu && !utils.Mobile() {
		showIOSMenu = false
	}

	mobile := utils.Mobile()
	if ok, _ := regexp.MatchString("(?i)(iPod|iPhone|iPad|Android)", r.UserAgent()); ok {
		mobile = true
	}

	ios := utils.IOS()
	if ok, _ := regexp.MatchString("(?i)(iPod|iPhone|iPad)", r.UserAgent()); ok {
		ios = true
	}

	android := utils.Android()
	if ok, _ := regexp.MatchString("(?i)(Android)", r.UserAgent()); ok {
		android = true
	}
	/*	key = strings.Replace(key, "\r", "\n", -1)
		key = strings.Replace(key, "\n\n", "\n", -1)
		key = strings.Replace(key, "\n", "\\\n", -1)*/

	setLang := r.FormValue("lang")

	funcMap := template.FuncMap{
		"noescape": func(s string) template.HTML {
			return template.HTML(s)
		},
	}

	data, err := static.Asset("static/index.html")
	t := template.New("template").Funcs(funcMap)
	t, err = t.Parse(string(data))
	if err != nil {
		log.Error("%v", err)
	}
	langs := ``
	if len(utils.LangList) > 0 {
		langs = strings.Join(utils.LangList, `,`)
	}
	b := new(bytes.Buffer)
	err = t.Execute(b, &index{
		DbOk:        true,
		Lang:        globalLangReadOnly[lang],
		Key:         r.FormValue(`key`),
		PKey:        r.FormValue(`pkey`),
		State:       r.FormValue(`state`),
		SetLang:     setLang,
		ShowIOSMenu: showIOSMenu,
		IOS:         ios,
		Accounts:    string(accounts),
		Thrust:      utils.Thrust,
		Android:     android,
		Mobile:      mobile,
		Langs:       langs,
		LogoExt:     utils.LogoExt,
		Version:     consts.VERSION})
	if err != nil {
		log.Error("%v", err)
	}
	w.Write(b.Bytes())
}
