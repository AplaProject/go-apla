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
	"io/ioutil"
	"path/filepath"
	"strings"
	//	"strconv"
	"encoding/json"

	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const NExportTpl = `export_tpl`

type exportInfo struct {
	//	Id   int    `json:"id"`
	Name   string `json:"name"`
	Global bool   `json:"global"`
}

type exportTplPage struct {
	Data      *CommonPage
	Message   string
	Contracts *[]exportInfo
	Pages     *[]exportInfo
	Tables    *[]exportInfo
	Menu      *[]exportInfo
}

func init() {
	newPage(NExportTpl)
}

func (c *Controller) getList(table, prefix string) (*[]exportInfo, error) {
	ret := make([]exportInfo, 0)
	contracts, err := c.GetAll(fmt.Sprintf(`select name from "%s_%s" order by name`, prefix, table), -1)
	if err != nil {
		return nil, err
	}
	global := prefix == `global`
	for _, ival := range contracts {
		//		id, _ := strconv.ParseInt(ival[`id`], 10, 32)
		ret = append(ret, exportInfo{ival["name"], global})
	}
	return &ret, nil
}

func (c *Controller) setVar(name, prefix string) (out string) {
	contracts := strings.Split(c.r.FormValue(name), `,`)
	if len(contracts) == 0 {
		return
	}
	out = `SetVar(`
	list := make([]string, 0)
	names := make([]string, 0)
	for _, icontract := range contracts {
		var state string
		icontract, _, state = getState(c.SessStateId, icontract)
		data, _ := c.Single(fmt.Sprintf(`select value from "%s_%s" where name=?`, state, name), icontract).String()
		//		fmt.Println(`Data`, err, data)
		if len(data) > 0 {
			names = append(names, prefix+`_`+icontract)
			if prefix == `p` {
				list = append(list, fmt.Sprintf("`%s_%s #= %s`", prefix, icontract, data))
			} else {
				list = append(list, fmt.Sprintf("`%s_%s = %s`", prefix, icontract, data))
			}
		}
	}
	out += strings.Join(list, ",\r\n") + ")\r\nTextHidden( " + strings.Join(names, ", ") + ")\r\n"
	return
}

func getState(stateId int64, name string) (out string, global int, state string) {
	state = utils.Int64ToStr(stateId)
	out = name
	if strings.HasPrefix(name, `global_`) {
		state = `global`
		global = 1
		out = out[len(`global_`):]
	}
	return
}

func (c *Controller) ExportTpl() (string, error) {
	name := c.r.FormValue(`name`)
	message := ``
	if len(name) > 0 {
		var out string
		tplname := filepath.Join(*utils.Dir, name+`.tpl`)
		out += `SetVar(
	global = 0,
	type_new_page_id = TxId(NewPage),
	type_new_menu_id = TxId(NewMenu),
	type_new_contract_id = TxId(NewContract),
	type_new_table_id = TxId(NewTable),	
	sc_conditions = "$citizen == #wallet_id#")
`
		out += c.setVar("smart_contracts", `sc`)
		out += c.setVar("pages", `p`)
		out += c.setVar("menu", `m`)
		//		out += c.setVar("tables", `t_`)

		out += "Json(`Head: \"" + c.r.FormValue(`title`) + "\",\r\n" + `Desc: "` + c.r.FormValue(`desc`) + `",
		Img: "/static/img/apps/ava.png",
		OnSuccess: {
			script: 'template',
			page: 'government',
			parameters: {}
		},
		TX: [`
		list := make([]string, 0)

		tables := strings.Split(c.r.FormValue("tables"), `,`)
		if len(tables) > 0 {
			for _, itable := range tables {
				if len(itable) == 0 {
					continue
				}
				var (
					state  string
					global int
				)
				itable, global, state = getState(c.SessStateId, itable)
				cols, _ := c.Single(fmt.Sprintf(`select columns_and_permissions->'update' from "%s_tables" where name=?`,
					state), itable).String()
				fmap := make(map[string]string)
				json.Unmarshal([]byte(cols), &fmap)
				fields := make([]string, 0)
				for key := range fmap {
					ikey := strings.ToLower(key)
					index := 0
					itype := ``
					if ok, _ := c.IsIndex(itable, ikey); ok {
						index = 1
					}
					coltype, _ := c.OneRow(`select data_type,character_maximum_length from information_schema.columns
where table_name = ? and column_name = ?`, itable, ikey).String()
					if len(coltype) > 0 {
						switch {
						case coltype[`data_type`] == "character varying":
							if coltype[`character_maximum_length`] == `32` {
								itype = "hash"
							} else {
								itype = `text`
							}
						case coltype[`data_type`] == `bigint`:
							itype = "int64"
						case strings.HasPrefix(coltype[`data_type`], `timestamp`):
							itype = "time"
						case strings.HasPrefix(coltype[`data_type`], `numeric`):
							itype = "money"
						case strings.HasPrefix(coltype[`data_type`], `double`):
							itype = "double"
						}
					}
					fields = append(fields, fmt.Sprintf(`["%s", "%s", "%d"]`, ikey, itype, index))
				}

				list = append(list, fmt.Sprintf(`{
		Forsign: 'global,table_name,columns',
		Data: {
			type: "NewTable",
			typeid: #type_new_table_id#,
			global: %d,
			table_name : "%s",
			columns: '[%s]',
			permissions: "$citizen == #wallet_id#"
			}
	   }`, global, itable[strings.IndexByte(itable, '_')+1:], strings.Join(fields, `,`)))
			}
		}
		contracts := strings.Split(c.r.FormValue("smart_contracts"), `,`)
		if len(contracts) > 0 {
			for _, icontract := range contracts {
				if len(icontract) == 0 {
					continue
				}
				var global int
				icontract, global, _ = getState(c.SessStateId, icontract)
				list = append(list, fmt.Sprintf(`{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: %d,
			name: "%s",
			value: $("#sc_%s").val(),
			conditions: $("#sc_conditions").val()
			}
	   }`, global, icontract, icontract))
			}
		}
		menu := strings.Split(c.r.FormValue("menu"), `,`)
		if len(menu) > 0 {
			for _, imenu := range menu {
				if len(imenu) == 0 {
					continue
				}
				var global int
				imenu, global, _ = getState(c.SessStateId, imenu)
				list = append(list, fmt.Sprintf(`{
		Forsign: 'global,name,value, conditions',
		Data: {
			type: "NewMenu",
			typeid: #type_new_menu_id#,
			name : "%s",
			value: $("#m_%s").val(),
			global: %d,
			conditions: "$citizen == #wallet_id#",
			}
	   }`, imenu, imenu, global))
			}
		}

		pages := strings.Split(c.r.FormValue("pages"), `,`)
		if len(pages) > 0 {
			for _, ipage := range pages {
				if len(ipage) == 0 {
					continue
				}
				var global int
				ipage, global, _ = getState(c.SessStateId, ipage)
				prefix := utils.Int64ToStr(c.SessStateId)
				if global == 1 {
					prefix = `global`
				}
				menu, _ := c.Single(fmt.Sprintf(`select menu from "%s_pages" where name=?`, prefix), ipage).String()
				if len(menu) == 0 {
					menu = "menu_default"
				}
				list = append(list, fmt.Sprintf(`{
		Forsign: 'global,name,value,menu,conditions',
		Data: {
			type: "NewPage",
			typeid: #type_new_page_id#,
			name : "%s",
			menu: "%s",
			value: $("#p_%s").val(),
			global: %d,
			conditions: "$citizen == #wallet_id#",
			}
	   }`, ipage, menu, ipage, global))
			}
		}
		out += strings.Join(list, ",\r\n") + "]`\r\n)"

		if err := ioutil.WriteFile(tplname, []byte(out), 0644); err != nil {
			message = err.Error()
		} else {
			message = fmt.Sprintf(`File %s has been created`, tplname)
		}
	}
	prefix := utils.Int64ToStr(c.SessStateId)
	loadlist := func(name string) (*[]exportInfo, error) {
		list, err := c.getList(name, prefix)
		if err != nil {
			return nil, err
		}
		glist, err := c.getList(name, `global`)
		if err != nil {
			return nil, err
		}
		*list = append(*list, *glist...)
		return list, nil
	}
	contracts, err := loadlist(`smart_contracts`)
	if err != nil {
		return ``, err
	}
	pages, err := loadlist(`pages`)
	if err != nil {
		return ``, err
	}
	tables, err := loadlist(`tables`)
	if err != nil {
		return ``, err
	}
	menu, err := loadlist(`menu`)
	if err != nil {
		return ``, err
	}

	//	fmt.Println(`Export`, contracts, pages, tables)
	pageData := exportTplPage{Data: c.Data, Contracts: contracts, Pages: pages, Tables: tables,
		Menu: menu, Message: message}
	return proceedTemplate(c, NExportTpl, &pageData)
}
