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
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

type showTablePage struct {
	Alert                 string
	Lang                  map[string]string
	WalletID              int64
	CitizenID             int64
	TxType                string
	TxTypeID              int64
	TimeNow               int64
	TableData             []map[string]string
	Columns               map[string]string
	ColumnsAndPermissions map[string]string
	TableName             string
	Global                string
}

// ShowTable shows data of the table
func (c *Controller) ShowTable() (string, error) {

	var err error

	var tableName string
	if utils.CheckInputData(c.r.FormValue("name"), "string") {
		tableName = c.r.FormValue("name")
	}

	global := c.r.FormValue("global")
	prefix := c.StateIDStr
	if global == "1" {
		prefix = "global"
	} else {
		global = "0"
	}
	t := &model.Table{}
	columns, err := t.GetColumnsAndPermissions(prefix, tableName)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	tableData, err := model.GetTableData(tableName, -1)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	for i, item := range tableData {
		for key, val := range item {
			if strings.HasPrefix(val, `data:image`) {
				tableData[i][key] = fmt.Sprintf(`<img src="%s">`, val)
			} else if len(val) == 20 && val[19] == 'Z' && val[10] == 'T' {
				tableData[i][key] = strings.Replace(val[:19], `T`, ` `, -1)
			} else if val == `NULL` {
				tableData[i][key] = ``
			} else if strings.IndexAny(val, "\x00\x01\x02\x03\x04\x05\x06") >= 0 {
				var out []byte
				for i, ch := range fmt.Sprintf(`%x`, val) {
					out = append(out, byte(ch))
					if (i & 1) == 1 {
						out = append(out, ' ')
					}
				}
				tableData[i][key] = string(out)
			} else {
				tableData[i][key] = strings.Replace(val, "\n", "\n<br>", -1)
			}
		}
	}

	TemplateStr, err := makeTemplate("show_table", "showTable", &showTablePage{
		Alert:     c.Alert,
		Lang:      c.Lang,
		Global:    global,
		WalletID:  c.SessWalletID,
		CitizenID: c.SessCitizenID,
		Columns:   columns,
		TableName: tableName,
		TableData: tableData})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
