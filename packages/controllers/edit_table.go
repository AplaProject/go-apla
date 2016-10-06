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
	"strings"
)

type editTablePage struct {
	Alert                 string
	SignData              string
	ShowSignData          bool
	CountSignArr          []int
	Lang                  map[string]string
	WalletId              int64
	CitizenId             int64
	TableName             string
	TxType                string
	TxTypeId              int64
	TimeNow               int64
	TableData             map[string]string
	Columns               map[string]string
	ColumnsAndPermissions map[string]string
	StateId               int64
	TablePermission       map[string]string
	Global                string
}

func (c *Controller) EditTable() (string, error) {

	var err error

	txType := "EditTable"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	var tableName string
	if utils.CheckInputData(c.r.FormValue("name"), "string") {
		tableName = c.r.FormValue("name")
	}

	s := strings.Split(tableName, "_")
	if len(s) < 2 {
		return "", utils.ErrInfo("incorrect table name")
	}
	prefix := s[0]
	if prefix != "global" && prefix != c.StateIdStr {
		return "", utils.ErrInfo("incorrect table name")
	}

	global := ""
	if prefix == "global" {
		global = "1"
	}

	tableData, err := c.OneRow(`SELECT * FROM "`+prefix+`_tables" WHERE name = ?`, tableName).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	tablePermission, err := c.GetMap(`SELECT data.* FROM "`+prefix+`_tables", jsonb_each_text(columns_and_permissions) as data WHERE name = ?`, "key", "value", tableName)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	columnsAndPermissions, err := c.GetMap(`SELECT data.* FROM "`+prefix+`_tables", jsonb_each_text(columns_and_permissions->'update') as data WHERE name = ?`, "key", "value", tableName)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	TemplateStr, err := makeTemplate("edit_table", "editTable", &editTablePage{
		Alert:                 c.Alert,
		Lang:                  c.Lang,
		ShowSignData:          c.ShowSignData,
		SignData:              "",
		WalletId:              c.SessWalletId,
		CitizenId:             c.SessCitizenId,
		CountSignArr:          c.CountSignArr,
		TableName:             tableName,
		TimeNow:               timeNow,
		TxType:                txType,
		TxTypeId:              txTypeId,
		StateId:               c.SessStateId,
		Global:                global,
		TablePermission:       tablePermission,
		ColumnsAndPermissions: columnsAndPermissions,
		TableData:             tableData})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
