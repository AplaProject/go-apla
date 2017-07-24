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
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

type editMenuPage struct {
	Alert     string
	Lang      map[string]string
	WalletID  int64
	CitizenID int64
	TxType    string
	TxTypeID  int64
	TimeNow   int64
	DataMenu  map[string]string
	StateID   int64
	Global    string
}

// EditMenu is a controller for editing menu
func (c *Controller) EditMenu() (string, error) {

	txType := "EditMenu"
	timeNow := time.Now().Unix()

	var err error

	global := c.r.FormValue("global")
	prefix := c.StateIDStr
	if global == "1" {
		prefix = "global"
	} else {
		global = "0"
	}

	var name string
	if utils.CheckInputData(c.r.FormValue("name"), "string") {
		name = c.r.FormValue("name")
	}

	menu := &model.Menu{}
	menu.SetTableName(prefix)
	err = menu.Get(name)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	dataMenu := menu.ToMap()
	if len(dataMenu[`conditions`]) == 0 {
		dataMenu[`conditions`] = "ContractConditions(`MainCondition`)"
	}

	TemplateStr, err := makeTemplate("edit_menu", "editMenu", &editMenuPage{
		Alert:     c.Alert,
		Lang:      c.Lang,
		Global:    global,
		WalletID:  c.SessWalletID,
		CitizenID: c.SessCitizenID,
		TimeNow:   timeNow,
		TxType:    txType,
		TxTypeID:  utils.TypeInt(txType),
		StateID:   c.SessStateID,
		DataMenu:  dataMenu})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
