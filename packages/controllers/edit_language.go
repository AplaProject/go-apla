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
	"encoding/json"
	"sort"

	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const nEditLanguage = `edit_language`

// LangRes contains the language resource
type LangRes struct {
	Lang string
	Text string
}

type editLanguagePage struct {
	Data     *CommonPage
	Name     string
	List     []LangRes
	Global   string
	TxType   string
	TxTypeID int64
	Unique   string
}

// ListLangRes is a slice of LangRes structures
type ListLangRes []LangRes

func init() {
	newPage(nEditLanguage)
}

func (a ListLangRes) Len() int           { return len(a) }
func (a ListLangRes) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ListLangRes) Less(i, j int) bool { return a[i].Lang < a[j].Lang }

// EditLanguage is a controller which edits language resources
func (c *Controller) EditLanguage() (string, error) {
	global := c.r.FormValue("global")
	prefix := "global"
	if global == "" || global == "0" {
		prefix = c.StateIDStr
		global = "0"
	}
	name := c.r.FormValue(`name`)

	txType := "NewLang"

	list := make([]LangRes, 0)
	if len(name) > 0 {
		language := &model.Language{}
		language.SetTablePrefix(prefix)
		err := language.Get(name)
		if err != nil {
			return "", err
		}
		res := language.Res
		var rmap map[string]string
		err = json.Unmarshal([]byte(res), &rmap)
		if err != nil {
			return "", err
		}
		for key, text := range rmap {
			list = append(list, LangRes{key, text})
		}
		sort.Sort(ListLangRes(list))
		txType = "EditLang"
	}
	pageData := editLanguagePage{Data: c.Data, Global: global, Name: name, List: list, TxType: txType,
		TxTypeID: utils.TypeInt(txType), Unique: ``}
	return proceedTemplate(c, nEditLanguage, &pageData)
}
