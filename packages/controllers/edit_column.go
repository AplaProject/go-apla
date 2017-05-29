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
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"strings"
)

type editColumnPage struct {
	Alert            string
	Lang             map[string]string
	WalletID         int64
	CitizenID        int64
	TxType           string
	TxTypeID         int64
	TimeNow          int64
	TableName        string
	StateID          int64
	ColumnPermission string
	ColumnName       string
	ColumnType       string
	CanIndex         bool
}

func (c *Controller) EditColumn() (string, error) {

	var err error

	txType := "EditColumn"
	txTypeID := utils.TypeInt(txType)
	timeNow := utils.Time()

	tableName := c.r.FormValue("tableName")
	columnName := c.r.FormValue("columnName")

	s := strings.Split(tableName, "_")
	if len(s) < 2 {
		return "", utils.ErrInfo("incorrect table name")
	}
	prefix := s[0]
	if prefix != "global" && prefix != c.StateIdStr {
		return "", utils.ErrInfo("incorrect table name")
	}

	columns, err := c.GetMap(`SELECT data.* FROM "`+prefix+`_tables", jsonb_each_text(columns_and_permissions->'update') as data WHERE name = ?`, "key", "value", tableName)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	TemplateStr, err := makeTemplate("edit_column", "editColumn", &editColumnPage{
		Alert:            c.Alert,
		Lang:             c.Lang,
		TableName:        tableName,
		ColumnName:       columnName,
		ColumnPermission: columns[columnName],
		ColumnType:       utils.GetColumnType(tableName, columnName),
		WalletID:         c.SessWalletID,
		CitizenID:        c.SessCitizenID,
		StateID:          c.SessStateID,
		TimeNow:          timeNow,
		TxType:           txType,
		TxTypeID:         txTypeID})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
