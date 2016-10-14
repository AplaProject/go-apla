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

package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/DayLightProject/go-daylight/packages/lib"
	"github.com/DayLightProject/go-daylight/packages/script"
	"github.com/DayLightProject/go-daylight/packages/smart"
	"github.com/DayLightProject/go-daylight/packages/static"
	"github.com/DayLightProject/go-daylight/packages/textproc"
	"github.com/russross/blackfriday"
)

type FieldInfo struct {
	Name     string `json:"name"`
	HtmlType string `json:"htmlType"`
	TxType   string `json:"txType"`
	Title    string `json:"title"`
	Value    string `json:"value"`
}

type FormCommon struct {
	//Lang   map[string]string
	/*	Address      string
		WalletId     int64
		CitizenId    int64
		StateId      int64
		StateName    string*/
	CountSignArr []byte
}

type FormInfo struct {
	TxName string
	Fields []FieldInfo
	Data   FormCommon
}

func init() {
	smart.Extend(&script.ExtendData{map[string]interface{}{
		"Balance":    Balance,
		"StateParam": StateParam,
		/*		"DBInsert":   DBInsert,
		 */
	}, map[string]string{
	//		`*parser.Parser`: `parser`,
	}})

	textproc.AddMaps(&map[string]textproc.MapFunc{`Table`: Table, `TxForm`: TxForm})
	textproc.AddFuncs(&map[string]textproc.TextFunc{`BtnEdit`: BtnEdit,
		`LiTemplate`: LiTemplate,
		`TemplateNav` : TemplateNav,
		`Title`:      Title, `MarkDown`: MarkDown, `Navigation`: Navigation, `PageTitle`: PageTitle,
		`PageEnd`: PageEnd})
}

// Reading and compiling contracts from smart_contracts tables
func LoadContracts() (err error) {
	var states []map[string]string
	prefix := []string{`global`}
	states, err = DB.GetAll(`select id from system_states order by id`, -1)
	if err != nil {
		return err
	}
	for _, istate := range states {
		prefix = append(prefix, istate[`id`])
	}
	for _, ipref := range prefix {
		if err = LoadContract(ipref); err != nil {
			return err
		}
	}
	return
}

// Reading and compiling contract of new state
func LoadContract(prefix string) (err error) {
	var contracts []map[string]string
	contracts, err = DB.GetAll(`select * from "`+prefix+`_smart_contracts" order by id`, -1)
	if err != nil {
		return err
	}
	for _, item := range contracts {
		if err = smart.Compile(item[`value`]); err != nil {
			return
		}
	}
	return
}

func Balance(wallet_id int64) (float64, error) {
	return DB.Single("SELECT amount FROM dlt_wallets WHERE wallet_id = ?", wallet_id).Float64()
}

func StateParam(idstate int64, name string) (string, error) {
	return DB.Single(`SELECT value FROM "`+Int64ToStr(idstate)+`_state_parameters" WHERE name = ?`, name).String()
}

func BtnEdit(vars *map[string]string, pars ...string) string {
	if len(pars) != 2 {
		return ``
	}
	return fmt.Sprintf(`<a type="button" class="btn btn-primary btn-block" 
	            onclick="load_page('%s', {id: %d, global: 0 } )"><i class="fa fa-cog"></i></a>`,
		pars[0], StrToInt64(pars[1]))
}

func Table(vars *map[string]string, pars *map[string]string) string {
	fields := `*`
	order := ``
	if val, ok := (*pars)[`Order`]; ok {
		order = `order by ` + lib.Escape(val)
	}
	if val, ok := (*pars)[`Fields`]; ok {
		fields = lib.Escape(val)
	}
	list, err := DB.GetAll(fmt.Sprintf(`select %s from %s %s`, fields,
		lib.EscapeName((*pars)[`Table`]), order), -1)
	if err != nil {
		return err.Error()
	}
	columns := textproc.Split((*pars)[`Columns`])
	out := `<table  class="table table-striped table-bordered table-hover"><tr>`
	for _, th := range *columns {
		out += `<th>` + th[0] + `</th>`
	}
	out += `</tr>`
	for _, item := range list {
		out += `<tr>`
		for key, value := range item {
			(*vars)[key] = value
		}
		for _, th := range *columns {
			val := textproc.Process(th[1], vars)
			if len(val) == 0 {
				val = textproc.Macro(th[1], vars)
			}
			out += `<td>` + val + `</td>`
		}
		out += `</tr>`
	}
	return out + `</table>`
}

func TxForm(vars *map[string]string, pars *map[string]string) string {
	return TXForm((*pars)[`Contract`])
}

func LiTemplate(vars *map[string]string, pars ...string) string {
	name := pars[0]
	title := name
	if len(pars) > 1 {
		title = pars[1]
	}
	return fmt.Sprintf(`<li><a href="#" onclick="load_template('%s'); HideMenu();"><span>%s</span></a></li>`,
		name, title)
}

func TemplateNav(vars *map[string]string, pars ...string) string {
	name := pars[0]
	title := name
	if len(pars) > 1 {
		title = pars[1]
	}
	return fmt.Sprintf(`<a href="#" onclick="load_template('%s'); HideMenu();"><span>%s</span></a>`,
		name, title)
}

func Navigation(vars *map[string]string, pars ...string) string {
	li := make([]string, 0)
	for _, ipar := range pars {
		li = append(li, ipar)
	}
	return textproc.Macro(fmt.Sprintf(`<ol class="breadcrumb"><span class="pull-right">
	<a href='#' onclick="load_page('editPage', {name: '#page#'} )">Edit</a></span>%s</ol>`,
		strings.Join(li, `&nbsp;/&nbsp;`)), vars)
}

func MarkDown(vars *map[string]string, pars ...string) string {
	return string(blackfriday.MarkdownCommon([]byte(pars[0])))
}

func Title(vars *map[string]string, pars ...string) string {
	return fmt.Sprintf(`<div class="content-heading">%s</div>`, pars[0])
}

func PageTitle(vars *map[string]string, pars ...string) string {
	return fmt.Sprintf(`<div class="panel panel-default"><div class="panel-heading"><div class="panel-title">%s</div></div><div class="panel-body">`, pars[0])
}

func PageEnd(vars *map[string]string, pars ...string) string {
	return `</div></div>`
}

func TXForm(name string) string {
	contract := smart.GetContract(name)
	if contract == nil || contract.Block.Info.(*script.ContractInfo).Tx == nil {
		return fmt.Sprintf(`there is not %s contract or parameters`, name)
	}
	funcMap := template.FuncMap{
		"sum": func(a, b interface{}) float64 {
			return InterfaceToFloat64(a) + InterfaceToFloat64(b)
		},
		"noescape": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
	data, err := static.Asset("static/tx_form.html")

	sign, err := static.Asset("static/signatures_new.html")
	if err != nil {
		return fmt.Sprint(err.Error())
	}

	t := template.New("template").Funcs(funcMap)
	t, err = t.Parse(string(data))
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	t = template.Must(t.Parse(string(sign)))

	b := new(bytes.Buffer)
	finfo := FormInfo{TxName: name, Fields: make([]FieldInfo, 0), Data: FormCommon{
		CountSignArr: []byte{1}}}
	for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
		if strings.Index(fitem.Tags, `map`) >= 0 {
			finfo.Fields = append(finfo.Fields, FieldInfo{Name: fitem.Name, HtmlType: "map",
				TxType: fitem.Type.String(), Title: fitem.Name})
		} else if strings.Index(fitem.Tags, `image`) >= 0 {
			finfo.Fields = append(finfo.Fields, FieldInfo{Name: fitem.Name, HtmlType: "image",
				TxType: fitem.Type.String(), Title: fitem.Name})
		} else if fitem.Type.String() == `string` {
			finfo.Fields = append(finfo.Fields, FieldInfo{Name: fitem.Name, HtmlType: "textinput",
				TxType: fitem.Type.String(), Title: fitem.Name})
		}

	}

	if err = t.Execute(b, finfo); err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	lines := strings.Split(b.String(), "\n")
	out := ``
	for _, line := range lines {
		if value := strings.TrimSpace(line); len(value) > 0 {
			out += value + "\r\n"
		}
	}
	return out
}
