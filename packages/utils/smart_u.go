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
	//	"reflect"
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

type PageTpl struct {
	Page     string
	Template string
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
	textproc.AddFuncs(&map[string]textproc.TextFunc{`BtnEdit`: BtnEdit, `Image`: Image,
		`LiTemplate`: LiTemplate, `LinkTemplate`: LinkTemplate, `BtnTemplate`: BtnTemplate,
		`TemplateNav`: TemplateNav,
		`Title`:       Title, `MarkDown`: MarkDown, `Navigation`: Navigation, `PageTitle`: PageTitle,
		`PageEnd`: PageEnd, `StateValue`: StateValue})
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

func LinkTemplate(vars *map[string]string, pars ...string) string {
	params := ``
	if len(pars) < 2 {
		return ``
	}
	if len(pars) >= 3 {
		params = pars[2]
	}
	return fmt.Sprintf(`<a onclick="load_template('%s', {%s} )">%s</a>`, pars[0], params, pars[1])
}

func BtnTemplate(vars *map[string]string, pars ...string) string {
	params := ``
	if len(pars) < 2 {
		return ``
	}
	if len(pars) >= 3 {
		params = pars[2]
	}
	return fmt.Sprintf(`<button type="button" class="btn btn-primary"onclick="load_template('%s', {%s} )">%s</button>`, pars[0], params, pars[1])
}

func Table(vars *map[string]string, pars *map[string]string) string {
	fields := `*`
	order := ``
	where := ``
	if val, ok := (*pars)[`Order`]; ok {
		order = `order by ` + lib.Escape(val)
	}
	if val, ok := (*pars)[`Where`]; ok {
		where = `where ` + lib.Escape(val)
	}
	if val, ok := (*pars)[`Fields`]; ok {
		fields = lib.Escape(val)
	}
	list, err := DB.GetAll(fmt.Sprintf(`select %s from %s %s %s`, fields,
		lib.EscapeName((*pars)[`Table`]), where, order), -1)
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
	return TXForm(vars, pars)
}

func Image(vars *map[string]string, pars ...string) string {
	alt := ``
	if len(pars) > 1 {
		alt = pars[1]
	}
	return fmt.Sprintf(`<img src="%s" alt="%s" style="display:block;">`, pars[0], alt)
}

func StateValue(vars *map[string]string, pars ...string) string {
	val, _ := StateParam(StrToInt64((*vars)[`state_id`]), pars[0])
	return val
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
	par1 := ""
	val1 := ""
	if len(pars) > 1 {
		par1 = pars[1]
	}
	if len(pars) > 2 {
		val1 = pars[2]
	}
	result := ""
	if len(par1) > 0 && len(val1) > 0 {
		result = fmt.Sprintf(`<a href="#" onclick="load_template('%s', {%s: '%s'}); HideMenu();"><span>%s</span></a>`,
			name, par1, val1, title)
	} else {
		result = fmt.Sprintf(`<a href="#" onclick="load_template('%s'); HideMenu();"><span>%s</span></a>`,
			name, title)
	}
	return result

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
	return fmt.Sprintf(`<div class="panel panel-default" data-sweet-alert><div class="panel-heading"><div class="panel-title">%s</div></div><div class="panel-body">`, pars[0])
}

func PageEnd(vars *map[string]string, pars ...string) string {
	return `</div></div>`
}

func TXForm(vars *map[string]string, pars *map[string]string) string {

	name := (*pars)[`Contract`]
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
		var value string
		if val, ok := (*vars)[fitem.Name]; ok {
			value = val
		}
		if strings.Index(fitem.Tags, `hidden`) >= 0 {
			continue
		}
		if strings.Index(fitem.Tags, `map`) >= 0 {
			finfo.Fields = append(finfo.Fields, FieldInfo{Name: fitem.Name, HtmlType: "map",
				TxType: fitem.Type.String(), Title: fitem.Name, Value: value})
		} else if strings.Index(fitem.Tags, `image`) >= 0 {
			finfo.Fields = append(finfo.Fields, FieldInfo{Name: fitem.Name, HtmlType: "image",
				TxType: fitem.Type.String(), Title: fitem.Name, Value: value})
		} else if fitem.Type.String() == `string` || fitem.Type.String() == `int64` {
			finfo.Fields = append(finfo.Fields, FieldInfo{Name: fitem.Name, HtmlType: "textinput",
				TxType: fitem.Type.String(), Title: fitem.Name, Value: value})
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

func ProceedTemplate(html string, data interface{}) (string, error) {

	defer func() {
		if r := recover(); r != nil {
			log.Error("proceedTemplate Recovered", r)
			fmt.Println(r)
		}
	}()
	pattern, err := static.Asset("static/" + html + ".html")
	if err != nil {
		return "", err
	}
	/*	funcMap := template.FuncMap{
			"makeCurrencyName": func(currencyId int64) string {
				if currencyId >= 1000 {
					return ""
				} else {
					return "d"
				}
			},
			"div": func(a, b interface{}) float64 {
				return utils.InterfaceToFloat64(a) / utils.InterfaceToFloat64(b)
			},
			"mult": func(a, b interface{}) float64 {
				return utils.InterfaceToFloat64(a) * utils.InterfaceToFloat64(b)
			},
			"round": func(a interface{}, num int) float64 {
				return utils.Round(utils.InterfaceToFloat64(a), num)
			},
			"len": func(s []map[string]string) int {
				return len(s)
			},
			"lenMap": func(s map[string]string) int {
				return len(s)
			},
			"sum": func(a, b interface{}) float64 {
				return utils.InterfaceToFloat64(a) + utils.InterfaceToFloat64(b)
			},
			"minus": func(a, b interface{}) float64 {
				return utils.InterfaceToFloat64(a) - utils.InterfaceToFloat64(b)
			},
			"js": func(s string) template.JS {
				return template.JS(s)
			},
			"join": func(s []string, sep string) string {
				return strings.Join(s, sep)
			},
			"strToInt64": func(text string) int64 {
				return utils.StrToInt64(text)
			},
			"strToInt": func(text string) int {
				return utils.StrToInt(text)
			},
			"bin2hex": func(text string) string {
				return string(utils.BinToHex([]byte(text)))
			},
			"int64ToStr": func(text int64) string {
				return utils.Int64ToStr(text)
			},
			"intToStr": func(text int) string {
				return utils.IntToStr(text)
			},
			"intToInt64": func(text int) int64 {
				return int64(text)
			},
			"rand": func() int {
				return utils.RandInt(0, 99999999)
			},
			"append": func(args ...interface{}) string {
				var result string
				for _, value := range args {
					switch value.(type) {
					case int64:
						result += utils.Int64ToStr(value.(int64))
					case float64:
						result += utils.Float64ToStr(value.(float64))
					case string:
						result += value.(string)
					}
				}
				return result
			},
			"replaceCurrency": func(text, name string) string {
				return strings.Replace(text, "[currency]", name, -1)
			},
			"replaceCurrencyName": func(text, name string) string {
				return strings.Replace(text, "[currency]", "D"+name, -1)
			},
			"cfCategoryLang": func(lang map[string]string, name string) string {
				return lang["cf_category_"+name]
			},
			"progressBarLang": func(lang map[string]string, name string) string {
				return lang["progress_bar_pct_"+name]
			},
			"checkProjectPs": func(ProjectPs map[string]string, id string) bool {
				if len(ProjectPs["ps"+id]) > 0 {
					return true
				} else {
					return false
				}
			},
			"cfPageTypeLang": func(lang map[string]string, name string) string {
				return lang["cf_"+name]
			},
			"notificationsLang": func(lang map[string]string, name string) string {
				return lang["notifications_"+name]
			},
			"issuffix": func(text, name string) bool {
				return strings.HasSuffix(text,name)
			},

		}
		t := template.Must(template.New("template").Funcs(funcMap).Parse(string(data)))*/
	//	t = template.Must(t.Parse(string(alert_success)))
	//	t = template.Must(t.Parse(string(signatures)))
	/*	t := template.New("template").Funcs(funcMap)
		t, err = t.Parse(string(data))
			if err != nil {
				w.Write([]byte(fmt.Sprintf("Error: %v", err)))
			}

			b := new(bytes.Buffer)
			err = t.Execute(b, c)
			if err != nil {
				w.Write([]byte(fmt.Sprintf("Error: %v", err)))
			}
			w.Write(b.Bytes())
	*/
	funcMap := template.FuncMap{
		"sum": func(a, b interface{}) int {
			//			fmt.Println(`TYPES`, reflect.TypeOf(a), reflect.TypeOf(b))
			return a.(int) + b.(int)
		},
		"noescape": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
	sign, err := static.Asset("static/signatures_new.html")
	if err != nil {
		return "", err
	}

	t := template.Must(template.New("template").Funcs(funcMap).Parse(string(pattern)))
	t = template.Must(t.Parse(string(sign)))
	b := new(bytes.Buffer)
	err = t.Execute(b, data)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
