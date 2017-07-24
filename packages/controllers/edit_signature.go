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

	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const nEditSignature = `edit_signature`

//SignRes contains the data of the signature
type SignRes struct {
	Param string `json:"name"`
	Text  string `json:"text"`
}

type editSignaturePage struct {
	Data       *CommonPage
	Name       string
	Conditions string
	Title      string
	List       []SignRes
	Global     string
	TxType     string
	TxTypeID   int64
	Unique     string
}

func init() {
	newPage(nEditSignature)
}

// EditSignature is a controller fo editing additional signatures
func (c *Controller) EditSignature() (string, error) {
	global := c.r.FormValue("global")
	prefix := "global"
	if global == "" || global == "0" {
		prefix = c.StateIDStr
		global = "0"
	}
	name := c.r.FormValue(`name`)

	txType := "NewSign"
	var title, cond string
	list := make([]SignRes, 0)
	if len(name) > 0 {
		signature := &model.Signatures{}
		signature.SetTableName(prefix)
		err := signature.Get(name)
		if err != nil {
			return "", err
		}
		var rmap map[string]interface{}
		cond = signature.Conditions
		err = json.Unmarshal([]byte(signature.Value), &rmap)
		if err != nil {
			return "", err
		}
		if val, ok := rmap[`title`]; ok {
			title = val.(string)
		}
		if val, ok := rmap[`params`]; ok {
			for _, item := range val.([]interface{}) {
				text := item.(map[string]interface{})
				list = append(list, SignRes{text[`name`].(string), text[`text`].(string)})
			}
		}
		txType = "EditSign"
	}
	pageData := editSignaturePage{Data: c.Data, List: list, Title: title, Conditions: cond,
		Global: global, Name: name, TxType: txType, TxTypeID: utils.TypeInt(txType), Unique: ``}
	return proceedTemplate(c, nEditSignature, &pageData)
}
