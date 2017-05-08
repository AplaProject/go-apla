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
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"strings"
	//"encoding/json"
	//"fmt"
)

type showTablePage struct {
	Alert                 string
	Lang                  map[string]string
	WalletId              int64
	CitizenId             int64
	TxType                string
	TxTypeId              int64
	TimeNow               int64
	TableData             []map[string]string
	Columns               map[string]string
	ColumnsAndPermissions map[string]string
	TableName             string
	Global                string
}

func (c *Controller) ShowTable() (string, error) {

	var err error

	var tableName string
	if utils.CheckInputData(c.r.FormValue("name"), "string") {
		tableName = c.r.FormValue("name")
	}

	global := c.r.FormValue("global")
	prefix := c.StateIdStr
	if global == "1" {
		prefix = "global"
	} else {
		global = "0"
	}
	var columns map[string]string
	columns, err = c.GetMap(`SELECT data.* FROM "`+prefix+`_tables", jsonb_each_text(columns_and_permissions->'update') as data WHERE name = ?`, "key", "value", tableName)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	//	columns["id"] = ""

	tableData, err := c.GetAll(`SELECT * FROM "`+tableName+`" order by id`, 1000)
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
		WalletId:  c.SessWalletId,
		CitizenId: c.SessCitizenId,
		Columns:   columns,
		//tableData : columnsAndPermissions,
		TableName: tableName,
		TableData: tableData})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
