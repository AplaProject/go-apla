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
	"github.com/DayLightProject/go-daylight/packages/utils"
	//"encoding/json"
	//"fmt"
)

type editPagePage struct {
	Alert        string
	SignData     string
	ShowSignData bool
	CountSignArr []int
	Lang         map[string]string
	WalletId     int64
	CitizenId    int64
	TxType       string
	TxTypeId     int64
	TimeNow      int64
	DataMenu     map[string]string
	DataPage     map[string]string
	AllMenu      []map[string]string
	StateId      int64
	Global       string
}

func (c *Controller) EditPage() (string, error) {

	txType := "EditPage"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	var err error

	global := c.r.FormValue("global")
	prefix := c.StateIdStr
	if global == "1" {
		prefix = "global"
	} else {
		global = "0"
	}

	var name string
	if utils.CheckInputData(c.r.FormValue("name"), "string") {
		name = c.r.FormValue("name")
	}

	dataPage, err := c.OneRow(`SELECT * FROM "`+prefix+`_pages" WHERE name = ?`, name).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	dataMenu, err := c.OneRow(`SELECT * FROM "`+prefix+`_menu" WHERE name = ?`, dataPage["menu"]).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	allMenu, err := c.GetAll(`SELECT * FROM "`+prefix+`_menu"`, -1)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	TemplateStr, err := makeTemplate("edit_page", "editPage", &editPagePage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		ShowSignData: c.ShowSignData,
		Global:       global,
		SignData:     "",
		WalletId:     c.SessWalletId,
		CitizenId:    c.SessCitizenId,
		CountSignArr: c.CountSignArr,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId,
		StateId:      c.SessStateId,
		AllMenu:      allMenu,
		DataMenu:     dataMenu,
		DataPage:     dataPage})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
