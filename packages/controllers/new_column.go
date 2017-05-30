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
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// NewColumn show the form for creating new column
func (c *Controller) NewColumn() (string, error) {

	var err error

	txType := "NewColumn"
	timeNow := utils.Time()

	tableName := lib.Escape(c.r.FormValue("tableName"))

	count, _ := c.NumIndexes(tableName)

	TemplateStr, err := makeTemplate("edit_column", "editColumn", &editColumnPage{
		Alert:            c.Alert,
		Lang:             c.Lang,
		TableName:        tableName,
		WalletID:         c.SessWalletID,
		CitizenID:        c.SessCitizenID,
		StateID:          c.SessStateID,
		ColumnName:       "",
		ColumnPermission: "",
		CanIndex:         count < consts.MAX_INDEXES,
		TimeNow:          timeNow,
		TxType:           txType,
		TxTypeID:         utils.TypeInt(txType)})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
