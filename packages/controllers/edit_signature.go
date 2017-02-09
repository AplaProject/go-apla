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
	//	"encoding/json"

	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const NEditSignature = `edit_signature`

type SignRes struct {
	Param string
	Text  string
}

type editSignaturePage struct {
	Data       *CommonPage
	Name       string
	Conditions string
	Title      string
	List       []SignRes
	Global     string
	TxType     string
	TxTypeId   int64
	Unique     string
}

func init() {
	newPage(NEditSignature)
}

func (c *Controller) EditSignature() (string, error) {
	global := c.r.FormValue("global")
	//	prefix := "global"
	if global == "" || global == "0" {
		//	prefix = c.StateIdStr
		global = "0"
	}
	name := c.r.FormValue(`name`)

	txType := "NewSign"

	//list := make([]LangRes, 0)
	if len(name) > 0 {
		/*		res, err := c.Single(`SELECT res FROM "`+prefix+`_languages" where name=?`, name).String()
				if err != nil {
					return "", err
				}
				var rmap map[string]string
				err = json.Unmarshal([]byte(res), &rmap)
				if err != nil {
					return "", err
				}
				for key, text := range rmap {
					list = append(list, LangRes{key, text})
				}
				sort.Sort(ListLangRes(list))*/
		txType = "EditLang"
	}
	txTypeId := utils.TypeInt(txType)
	pageData := editSignaturePage{Data: c.Data, Global: global, Name: name, TxType: txType, TxTypeId: txTypeId, Unique: ``}
	return proceedTemplate(c, NEditSignature, &pageData)
}
