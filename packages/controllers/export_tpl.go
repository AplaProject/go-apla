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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"regexp"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const nExportTpl = `export_tpl`

// ExpContract contains information about a contract
type ExpContract struct {
	Contract string
	Global   int
	Name     string
}

// ExpSlice is a slice of ExpContract
type ExpSlice []ExpContract

func (a ExpSlice) Len() int      { return len(a) }
func (a ExpSlice) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ExpSlice) Less(i, j int) bool {
	iused := smart.GetUsedContracts(a[i].Name, uint32(0), true)
	if iused == nil {
		return true
	}
	for _, val := range iused {
		if val == a[j].Name {
			return false
		}
	}
	return true
}

type exportInfo struct {
	//	Id   int    `json:"id"`
	Name   string `json:"name"`
	Global bool   `json:"global"`
}

type exportTplPage struct {
	Data       *CommonPage
	Message    string
	Contracts  *[]exportInfo
	Pages      *[]exportInfo
	Tables     *[]exportInfo
	DataTables *[]exportInfo
	Menu       *[]exportInfo
	Params     *[]exportInfo
}

func init() {
	newPage(nExportTpl)
}

func (c *Controller) getList(table, prefix string) (*[]exportInfo, error) {
	ret := make([]exportInfo, 0)
	contracts, err := model.GetNameList(fmt.Sprintf("%s_%s", prefix, table), -1)
	if err != nil {
		return nil, err
	}
	global := prefix == `global`
	for _, ival := range contracts {
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
		icontract, _, state = getState(c.SessStateID, icontract)
		data, _ := model.GetConditionsAndValue(fmt.Sprintf("%s_%s", state, name), icontract)
		if len(data) > 0 && len(data[`value`]) > 0 {
			names = append(names, prefix+`_`+icontract)
			list = append(list, fmt.Sprintf("`%s_%s #= %s`", prefix, icontract, strings.Replace(data[`value`], "`", "``", -1)))
			names = append(names, prefix+`c_`+icontract)
			list = append(list, fmt.Sprintf("`%sc_%s #= %s`", prefix, icontract, strings.Replace(data[`conditions`], "`", `"`, -1)))
		}
	}
	out += strings.Join(list, ",\r\n") + ")\r\nTextHidden( " + strings.Join(names, ", ") + ")\r\n"
	return
}

func (c *Controller) setData(name, prefix string) (out string) {
	datatables := strings.Split(c.r.FormValue(name), `,`)
	if len(datatables) == 0 {
		return
	}
	out = `SetVar(`
	list := make([]string, 0)
	names := make([]string, 0)

	for _, itable := range datatables {
		if len(itable) == 0 {
			continue
		}
		var (
			state, tblname string
			global         int
		)
		tblname = itable[strings.IndexByte(itable, '_')+1:]
		itable, global, state = getState(c.SessStateID, itable)
		contname := fmt.Sprintf(`Export%d_%s`, global, tblname)
		fmt.Println(itable, global, state)
		if global == 1 {
			tblname = fmt.Sprintf(`"global_%s"`, tblname)
		} else {
			tblname = fmt.Sprintf(`Table("%s")`, tblname)
		}
		data, _ := model.GetTableData(itable, -1)
		if len(data) == 0 {
			continue
		}
		pars := make([]string, 0)
		lines := make([]string, 0)
		null := make(map[string]string)
		for key := range data[0] {
			if key != `rb_id` && key != `id` {
				pars = append(pars, key)
				coltype, _ := model.GetColumnDataTypeCharMaxLength(itable, key)
				if len(coltype) > 0 {
					ival := `0`
					switch {
					case coltype[`data_type`] == "character varying", coltype[`data_type`] == "bytea":
						ival = ``
					case strings.HasPrefix(coltype[`data_type`], `timestamp`):
						ival = "NULL"
					}
					null[key] = ival
				}
			}
		}
		contract := fmt.Sprintf(`contract %s {
func action {
	var tblname, fields string
	tblname = %s
	fields = "%s"
`, contname, tblname, strings.Join(pars, `,`))
		for _, ilist := range data {
			params := make([]string, 0)
			for _, ipar := range pars {
				val := ilist[ipar]
				if strings.IndexByte(val, 0) >= 0 {
					val = `wrong parameter`
				}
				if val == `NULL` {
					val = null[ipar]
				}
				params = append(params, fmt.Sprintf(`"%s"`, converter.EscapeForJSON(val)))
			}
			lines = append(lines, fmt.Sprintf(`	DBInsert(tblname, fields, %s)`, strings.Join(params, `,`)))
		}
		contract += strings.Join(lines, "\r\n")
		contract += `
	}
}`
		names = append(names, prefix+`_`+contname)
		list = append(list, fmt.Sprintf("`%s_%s #= %s`", prefix, contname, contract))
	}
	out += strings.Join(list, ",\r\n") + ")\r\nTextHidden( " + strings.Join(names, ", ") + ")\r\n"
	return
}

func (c *Controller) setAppend(name, prefix string) (out string) {
	inlist := make([]string, 0)
	json.Unmarshal([]byte(c.r.FormValue(`app`+name)), &inlist)
	if len(inlist) == 0 {
		return
	}
	out = `SetVar(`
	names := make([]string, 0)
	list := make([]string, 0)
	for _, ilist := range inlist {
		//		var state string

		lr := strings.SplitN(ilist, `=`, 2)
		iname, _, _ := getState(c.SessStateID, lr[0])
		if len(lr) > 1 {
			names = append(names, prefix+`_`+iname)
			list = append(list, fmt.Sprintf("`%s_%s #= %s`", prefix, iname, lr[1]))
		}
	}
	out += strings.Join(list, ",\r\n") + ")\r\nTextHidden( " + strings.Join(names, ", ") + ")\r\n"
	return
}

func (c *Controller) setLang() (out string) {
	out = "SetVar(`l_lang #= "
	list := make(map[string]string)
	lang := &model.Language{}
	res, _ := lang.GetAll(converter.Int64ToStr(c.SessStateID))
	for _, ires := range res {
		list[ires.Name] = ires.Res
	}
	val, _ := json.Marshal(list)
	out += string(val) + "`)\r\nTextHidden(l_lang)\r\n"
	return
}

func getState(stateID int64, name string) (out string, global int, state string) {
	state = converter.Int64ToStr(stateID)
	out = name
	if strings.HasPrefix(name, `global_`) {
		state = `global`
		global = 1
		out = out[len(`global_`):]
	}
	return
}

// ExportTpl is a handle function which can export different information
func (c *Controller) ExportTpl() (string, error) {
	name := c.r.FormValue(`name`)
	message := ``
	if len(name) > 0 {
		var out string
		tplname := filepath.Join(*utils.Dir, name+`.tpl`)
		out += `SetVar(
	global = 0,
	type_new_page_id = TxId(NewPage),
	type_append_page_id = TxId(AppendPage),
	type_new_menu_id = TxId(NewMenu),
	type_edit_table_id = TxId(EditTable),
	type_edit_column_id = TxId(EditColumn),
	type_append_menu_id = TxId(AppendMenu),
	type_new_lang_id = TxId(NewLang),
	type_new_contract_id = TxId(NewContract),
	type_activate_contract_id = TxId(ActivateContract),
	type_new_sign_id = TxId(NewSign),
	type_new_state_params_id = TxId(NewStateParameters), 
	type_new_table_id = TxId(NewTable))
`
		out += c.setVar("smart_contracts", `sc`)

		contracts := strings.Split(c.r.FormValue("smart_contracts"), `,`)

		signlist := make(map[string]bool, 0)

		if len(contracts) > 0 {
			for _, icontract := range contracts {
				if len(icontract) == 0 {
					continue
				}
				var global int
				icontract, global, _ = getState(c.SessStateID, icontract)
				state := c.SessStateID
				if global == 1 {
					state = 0
				}
				contract := smart.GetContract(icontract, uint32(state))
				if contract.Block.Info.(*script.ContractInfo).Tx != nil {
					signs := `SetVar(`
					names := make([]string, 0)
					list := make([]string, 0)
					for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
						if strings.Index(fitem.Tags, `signature`) >= 0 {
							if ret := regexp.MustCompile(`(?is)signature:([\w_\d]+)`).FindStringSubmatch(fitem.Tags); len(ret) == 2 {
								pref := converter.Int64ToStr(state)
								if state == 0 {
									pref = `global`
								}
								sign := &model.Signature{}
								sign.SetTablePrefix(pref)
								err := sign.Get(ret[1])
								if err != nil {
									break
								}
								names = append(names, `sign_`+ret[1])
								list = append(list, fmt.Sprintf("`sign_%s #= %s`", ret[1], strings.Replace(sign.Value, "`", `"`, -1)))
								names = append(names, `signc_`+ret[1])
								list = append(list, fmt.Sprintf("`signc_%s #= %s`", ret[1], strings.Replace(sign.Conditions, "`", `"`, -1)))
								signlist[fmt.Sprintf(`%d%s`, global, ret[1])] = true
							}
						}
					}
					if len(list) > 0 {
						signs += strings.Join(list, ",\r\n") + ")\r\nTextHidden( " + strings.Join(names, ", ") + ")\r\n"
						out += signs
					}
				}
			}
		}
		out += c.setVar("pages", `p`)
		out += c.setVar("menu", `m`)
		out += c.setVar("state_parameters", `pa`)
		out += c.setData("datatables", `d`)
		out += c.setAppend("pages", `ap`)
		out += c.setAppend("menu", `am`)
		if c.r.FormValue(`lang`) == `lang` {
			out += c.setLang()
		}
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
				itable, global, state = getState(c.SessStateID, itable)
				t := &model.Table{}
				cols, _ := t.GetColumnsAndPermissions(state, itable)
				fields := make([]string, 0)
				for key := range cols {
					ikey := strings.ToLower(key)
					index := 0
					itype := ``
					if ok, _ := model.IsIndex(itable, ikey); ok {
						index = 1
					}
					coltype, _ := model.GetColumnDataTypeCharMaxLength(itable, ikey)
					if len(coltype) > 0 {
						switch {
						case coltype[`data_type`] == "character varying":
							itype = `text`
						case coltype[`data_type`] == "bytea":
							itype = "hash"
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
							columns: '[%s]'
							}
					   }`, global, itable[strings.IndexByte(itable, '_')+1:], strings.Join(fields, `,`)))
				table := &model.Table{}
				table.SetTablePrefix(state)
				table.Get(itable)
				var jperm map[string]interface{}
				json.Unmarshal([]byte(table.Permissions), &jperm)
				var toedit bool
				vals := make(map[string]string)
				re, _ := regexp.Compile(`^\$citizen\s*==\s*-?\d+$`)
				for _, val := range []string{`insert`, `new_column`, `general_update`} {
					if !re.MatchString(jperm[val].(string)) {
						toedit = true
						vals[val] = jperm[val].(string)
					} else {
						vals[val] = "$citizen == #wallet_id#"
					}
				}
				tablepref := `#state_id#`
				if state == `global` {
					tablepref = state
				}
				if toedit {
					list = append(list, fmt.Sprintf(`{
		Forsign: 'table_name,general_update,insert,new_column',
		Data: {
			type: "EditTable",
			typeid: #type_edit_table_id#,
			table_name : "%s_%s",
			general_update: "%s",
			insert: "%s",
			new_column: "%s",
			}
	   }`, tablepref, itable[strings.IndexByte(itable, '_')+1:], converter.EscapeForJSON(vals[`general_update`]),
						converter.EscapeForJSON(vals[`insert`]), converter.EscapeForJSON(vals[`new_column`])))
				}
				jpermUpdate := jperm["update"].(map[string]interface{})
				for key, field := range jpermUpdate {
					if !re.MatchString(field.(string)) {
						list = append(list, fmt.Sprintf(`{
		Forsign: 'table_name,column_name,permissions',
		Data: {
			type: "EditColumn",
			typeid: #type_edit_column_id#,
			table_name : "%s_%s",
			column_name: "%s",
			permissions: "%s",
			}
	   }`, tablepref, itable[strings.IndexByte(itable, '_')+1:], key, converter.EscapeForJSON(field.(string))))
					}
				}
			}
		}

		datatables := strings.Split(c.r.FormValue("datatables"), `,`)
		if len(datatables) > 0 {
			for _, itable := range datatables {
				if len(itable) == 0 {
					continue
				}
				var (
					tblname string
					global  int
				)
				tblname = itable[strings.IndexByte(itable, '_')+1:]
				itable, global, _ = getState(c.SessStateID, itable)
				contname := fmt.Sprintf(`Export%d_%s`, global, tblname)

				list = append(list, fmt.Sprintf(`{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: %d,
			name: "%s",
			value: $("#d_%s").val(),
			conditions: "ContractConditions(\"MainCondition\")"
			}
	   }`, global, contname, contname))

				list = append(list, fmt.Sprintf(`{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: %d,
			id: "%s"
			}
	   }`, global, contname))

				list = append(list, fmt.Sprintf(`{
				Forsign: '',
				Data: {
					type: "Contract",
					global: %d,
					name: "%s"
					}
			}`, global, contname))
			}
		}

		cont := strings.Split(c.r.FormValue("smart_contracts"), `,`)
		if len(cont) > 0 {
			contracts := make(ExpSlice, 0)
			for _, icontract := range cont {
				var global int

				if len(icontract) == 0 {
					continue
				}
				icontract, global, _ = getState(c.SessStateID, icontract)
				var name string
				if global > 0 {
					name = `@0` + icontract
				} else {
					name = fmt.Sprintf(`@%d%s`, c.SessStateID, icontract)
				}
				contracts = append(contracts, ExpContract{Contract: icontract, Global: global,
					Name: name})
			}
			//			sort.Slice(contracts, sortContracts) for golang >= ver 1.8
			sort.Sort(contracts)
			for _, icont := range contracts {
				global := icont.Global
				icontract := icont.Contract
				//				icontract, global, _ = getState(c.SessStateID, icontract)
				list = append(list, fmt.Sprintf(`{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewContract",
			typeid: #type_new_contract_id#,
			global: %d,
			name: "%s",
			value: $("#sc_%[2]s").val(),
			conditions: $("#scc_%[2]s").val()
			}
	   }`, global, icontract))
				list = append(list, fmt.Sprintf(`{
		Forsign: 'global,id',
		Data: {
			type: "ActivateContract",
			typeid: #type_activate_contract_id#,
			global: %d,
			id: "%s"
			}
	   }`, global, icontract))
			}
		}
		for signitem := range signlist {
			list = append(list, fmt.Sprintf(`{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewSign",
			typeid: #type_new_sign_id#,
			global: %s,
			name: "%s",
			value: $("#sign_%[2]s").val(),
			conditions: $("#signc_%[2]s").val()
			}
	   }`, signitem[:1], signitem[1:]))
		}
		params := strings.Split(c.r.FormValue("state_parameters"), `,`)
		if len(params) > 0 {
			for _, iparam := range params {
				if len(iparam) == 0 {
					continue
				}
				//				var global int
				iparam, _, _ = getState(c.SessStateID, iparam)
				list = append(list, fmt.Sprintf(`{
		Forsign: 'name,value,conditions',
		Data: {
			type: "NewStateParameters",
			typeid: #type_new_state_params_id#,
			name : "%s",
			value: $("#pa_%[1]s").val(),
			conditions: $("#pac_%[1]s").val(),
			}
	   }`, iparam))
			}
		}

		menu := strings.Split(c.r.FormValue("menu"), `,`)
		if len(menu) > 0 {
			for _, imenu := range menu {
				if len(imenu) == 0 {
					continue
				}
				var global int
				imenu, global, _ = getState(c.SessStateID, imenu)
				list = append(list, fmt.Sprintf(`{
		Forsign: 'global,name,value,conditions',
		Data: {
			type: "NewMenu",
			typeid: #type_new_menu_id#,
			name : "%s",
			value: $("#m_%[1]s").val(),
			global: %d,
			conditions: $("#mc_%[1]s").val()
			}
	   }`, imenu, global))
			}
		}

		pages := strings.Split(c.r.FormValue("pages"), `,`)
		if len(pages) > 0 {
			for _, ipage := range pages {
				if len(ipage) == 0 {
					continue
				}
				var global int
				ipage, global, _ = getState(c.SessStateID, ipage)
				prefix := converter.Int64ToStr(c.SessStateID)
				if global == 1 {
					prefix = `global`
				}
				page := &model.Page{}
				page.SetTablePrefix(prefix)
				page.Get(ipage)
				menu := page.Menu
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
			value: $("#p_%[1]s").val(),
			global: %[3]d,
			conditions: $("#pc_%[1]s").val(),
			}
	   }`, ipage, menu, global))
			}
		}
		langs := strings.Split(c.r.FormValue("lang"), `,`)
		if len(langs) > 0 && langs[0] == `lang` {
			list = append(list, `{
				Forsign: 'name,trans',
				Data: {
					type: "NewLang",
					typeid: #type_new_lang_id#,
					name : "",
					trans: $("#l_lang").val(),
					}
				}`)
		}

		inlist := make([]string, 0)
		json.Unmarshal([]byte(c.r.FormValue(`apppages`)), &inlist)
		if len(inlist) >= 0 {
			for _, ilist := range inlist {
				var global int
				var iname string

				lr := strings.SplitN(ilist, `=`, 2)
				iname, global, _ = getState(c.SessStateID, lr[0])
				if len(lr) > 1 {
					list = append(list, fmt.Sprintf(`{
			Forsign: 'global,name,value',
			Data: {
				type: "AppendPage",
				typeid: #type_append_page_id#,
				name : "%s",
				value: $("#ap_%[1]s").val(),
				global: %d
				}
		}`, iname, global))
				}
			}
		}
		json.Unmarshal([]byte(c.r.FormValue(`appmenu`)), &inlist)
		if len(inlist) >= 0 {
			for _, ilist := range inlist {
				var global int
				var iname string

				lr := strings.SplitN(ilist, `=`, 2)
				iname, global, _ = getState(c.SessStateID, lr[0])
				if len(lr) > 1 {
					list = append(list, fmt.Sprintf(`{
			Forsign: 'global,name,value',
			Data: {
				type: "AppendMenu",
				typeid: #type_append_menu_id#,
				name : "%s",
				value: $("#am_%[1]s").val(),
				global: %d
				}
		}`, iname, global))
				}
			}
		}

		out += strings.Replace(strings.Join(list, ",\r\n"), "`", `\"`, -1) + "]`\r\n)"

		if err := ioutil.WriteFile(tplname, []byte(out), 0644); err != nil {
			message = err.Error()
		} else {
			message = fmt.Sprintf(`File %s has been created`, tplname)
		}
	}
	prefix := converter.Int64ToStr(c.SessStateID)
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
	params, err := c.getList(`state_parameters`, prefix)
	if err != nil {
		return ``, err
	}
	//	fmt.Println(`Export`, contracts, pages, tables)
	pageData := exportTplPage{Data: c.Data, Contracts: contracts, Pages: pages, Tables: tables,
		Menu: menu, Params: params, Message: message}
	return proceedTemplate(c, nExportTpl, &pageData)
}
