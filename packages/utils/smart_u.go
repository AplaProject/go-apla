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
	"strconv"
	"strings"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/static"
	"github.com/EGaaS/go-egaas-mvp/packages/textproc"
	"github.com/russross/blackfriday"
	"github.com/shopspring/decimal"
)

type FieldInfo struct {
	Name     string      `json:"name"`
	HtmlType string      `json:"htmlType"`
	TxType   string      `json:"txType"`
	Title    string      `json:"title"`
	Value    interface{} `json:"value"`
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
	TxName    string
	Unique    template.JS
	OnSuccess template.JS
	Fields    []FieldInfo
	Data      FormCommon
}

type TxInfo struct {
	Name     string `json:"name"`
	Id       string `json:"id"`
	HtmlType string `json:"htmlType"`
}

type TxButtonInfo struct {
	TxName    string
	Unique    template.JS
	OnSuccess template.JS
	Fields    []TxInfo
	Data      FormCommon
}

type CommonPage struct {
	Address      string
	WalletId     int64
	CitizenId    int64
	StateId      int64
	StateName    string
	CountSignArr []int
}

type PageTpl struct {
	Page     string
	Template string
	Unique   string
	Data     *CommonPage
}

type SelList struct {
	Cur  int64          `json:"cur"`
	List map[int]string `json:"list"`
}

type SelInfo struct {
	Id   int64
	Name string
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

	textproc.AddMaps(&map[string]textproc.MapFunc{`Table`: Table, `TxForm`: TxForm, `TxButton`: TXButton})
	textproc.AddFuncs(&map[string]textproc.TextFunc{`Address`: IdToAddress, `BtnEdit`: BtnEdit,
		`Image`: Image, `Div`: Div, `P`: P, `Em`: Em, `Small`: Small, `Divs`: Divs, `DivsEnd`: DivsEnd,
		`LiTemplate`: LiTemplate, `LinkTemplate`: LinkTemplate, `BtnTemplate`: BtnTemplate, `BtnSys`: BtnSys,
		`AppNav`: AppNav, `TemplateNav`: TemplateNav, `SysLink`: SysLink, `CmpTime`: CmpTime,
		`Title`: Title, `MarkDown`: MarkDown, `Navigation`: Navigation, `PageTitle`: PageTitle,
		`PageEnd`: PageEnd, `StateValue`: StateValue, `Json`: JsonScript, `And`: And, `Or`: Or,
		`TxId`: TxId, `SetVar`: SetVar, `GetRow`: GetRowVars, `GetOne`: GetOne, `TextHidden`: TextHidden,
		`ValueById`: ValueById, `FullScreen`: FullScreen, `Ring`: Ring, `WiBalance`: WiBalance,
		`WiAccount`: WiAccount, `WiCitizen`: WiCitizen, `Map`: Map, `MapPoint`: MapPoint, `StateLink`: StateLink,
		`If`: If, `Func`: Func, `Date`: Date, `DateTime`: DateTime, `Now`: Now, `Input`: Input,
		`Form`: Form, `FormEnd`: FormEnd, `Label`: Label, `Select`: Select, `Param`: Param, `Mult`: Mult,
	})
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
	LoadContract(`global`)
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
		if err = smart.Compile(item[`value`], prefix); err != nil {
			return
		}
	}
	return
}

func Balance(wallet_id int64) (decimal.Decimal, error) {
	balance, err := DB.Single("SELECT amount FROM dlt_wallets WHERE wallet_id = ?", wallet_id).String()
	if err != nil {
		return decimal.New(0, 0), err
	}
	return decimal.NewFromString(balance)
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

func Param(vars *map[string]string, pars ...string) string {
	if val, ok := (*vars)[pars[0]]; ok {
		return val
	}
	return ``
}

func ifValue(val string) bool {
	var sep string
	for _, item := range []string{`==`, `!=`, `<=`, `>=`, `<`, `>`} {
		if strings.Index(val, item) >= 0 {
			sep = item
			break
		}
	}
	cond := []string{val}
	if len(sep) > 0 {
		cond = strings.SplitN(val, sep, 2)
		cond[0], cond[1] = strings.Trim(cond[0], `"`), strings.Trim(cond[1], `"`)
	}
	switch sep {
	case ``:
		return len(val) > 0 && val != `0` && val != `false`
	case `==`:
		return len(cond) == 2 && strings.TrimSpace(cond[0]) == strings.TrimSpace(cond[1])
	case `!=`:
		return len(cond) == 2 && strings.TrimSpace(cond[0]) != strings.TrimSpace(cond[1])
	case `>`, `<`, `<=`, `>=`:
		ret0, _ := decimal.NewFromString(cond[0])
		ret1, _ := decimal.NewFromString(cond[1])
		if len(cond) == 2 {
			var bin bool
			if sep == `>` || sep == `<=` {
				bin = ret0.Cmp(ret1) > 0
			} else {
				bin = ret0.Cmp(ret1) < 0
			}
			if sep == `<=` || sep == `>=` {
				bin = !bin
			}
			return bin
		}
	}
	return false
}

func And(vars *map[string]string, pars ...string) string {
	for _, item := range pars {
		if !ifValue(item) {
			return `0`
		}
	}
	return `1`
}

func Or(vars *map[string]string, pars ...string) string {
	for _, item := range pars {
		if ifValue(item) {
			return `1`
		}
	}
	return `0`
}

func CmpTime(vars *map[string]string, pars ...string) string {
	if len(pars) < 2 {
		return ``
	}
	prepare := func(val string) string {
		val = strings.Replace(val, `T`, ` `, -1)
		if len(val) > 19 {
			val = val[:19]
		}
		return val
	}
	left := prepare(pars[0])
	right := prepare(pars[1])
	if left == right {
		return `0`
	}
	if left < right {
		return `-1`
	}
	return `1`
}

func If(vars *map[string]string, pars ...string) string {
	if len(pars) < 2 {
		return ``
	}
	if ifValue(pars[0]) {
		return pars[1]
	}
	if len(pars) > 2 {
		return pars[2]
	}
	return ``
}

func Now(vars *map[string]string, pars ...string) string {
	var (
		cut   int
		query string
	)
	if len(pars) == 0 || pars[0] == `` {
		query = `select round(extract(epoch from now()))::integer`
		cut = 10
	} else {
		query = `select now()`
		switch pars[0] {
		case `datetime`:
			cut = 19
		default:
			query = fmt.Sprintf(`select to_char(now(), '%s')`, pars[0])
		}
	}
	ret, err := DB.Single(query).String()
	if err != nil {
		return err.Error()
	}
	if cut > 0 {
		ret = strings.Replace(ret[:cut], `T`, ` `, -1)
	}
	return ret
}

func Input(vars *map[string]string, pars ...string) string {
	var (
		class, value, placeholder string
	)
	itype := `text`
	if len(pars) > 1 {
		class = pars[1]
	}
	if len(pars) > 2 {
		placeholder = LangText(pars[2], int(StrToInt64((*vars)[`state_id`])), (*vars)[`accept_lang`])
	}
	if len(pars) > 3 {
		itype = pars[3]
	}
	if len(pars) > 4 {
		value = pars[4]
	}
	return fmt.Sprintf(`<input type="%s" id="%s" placeholder="%s" class="%s" value="%s">`,
		itype, pars[0], placeholder, class, value)
}

func Func(vars *map[string]string, pars ...string) string {
	if len(pars) == 0 {
		return ``
	}
	return strings.TrimSpace(pars[0])
}

func JsonScript(vars *map[string]string, pars ...string) string {
	if len(pars) == 0 {
		return ``
	}
	return fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
	var jdata = { 
%s 
}
</script>`, pars[0])
}

func FullScreen(vars *map[string]string, pars ...string) string {
	wide := `add`
	if len(pars) > 0 && pars[0] == `0` {
		wide = `remove`
	}
	return fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
	$("body").%sClass('wide');
</script>`, wide)
}

func GetRowVars(vars *map[string]string, pars ...string) string {
	if len(pars) != 4 && len(pars) != 3 {
		return ``
	}
	where := ``
	if len(pars) == 4 {
		where = ` where ` + lib.EscapeName(pars[2]) + `='` + lib.Escape(pars[3]) + `'`
	} else if len(pars) == 3 {
		where = ` where ` + lib.Escape(pars[2])
	}
	value, err := DB.OneRow(`select * from ` + lib.EscapeName(pars[1]) + where).String()
	if err != nil {
		return err.Error()
	}
	for key, val := range value {
		(*vars)[pars[0]+`_`+key] = val
	}
	return ``
}

func GetOne(vars *map[string]string, pars ...string) string {
	if len(pars) < 2 {
		return ``
	}
	where := ``
	if len(pars) == 4 {
		where = ` where ` + lib.EscapeName(pars[2]) + `='` + lib.Escape(pars[3]) + `'`
	} else if len(pars) == 3 {
		where = ` where ` + lib.Escape(pars[2])
	}
	value, err := DB.Single(`select ` + lib.Escape(pars[0]) + ` from ` + lib.EscapeName(pars[1]) + where).String()
	if err != nil {
		return err.Error()
	}
	return strings.Replace(value, "\n", "\n<br>", -1)
}

func getClass(class string) string {
	list := strings.Split(class, ` `)
	for i, ilist := range list {
		if strings.HasPrefix(ilist, `xs-`) || strings.HasPrefix(ilist, `sm-`) ||
			strings.HasPrefix(ilist, `md-`) || strings.HasPrefix(ilist, `lg`) {
			list[i] = `col-` + ilist
		}
	}
	return strings.Join(list, ` `)
}

func getTag(tag string, pars ...string) (out string) {
	if len(pars) == 0 {
		return
	}
	out = fmt.Sprintf(`<%s class="%s">`, tag, getClass(pars[0]))
	for i := 1; i < len(pars); i++ {
		out += pars[i]
	}
	return out + fmt.Sprintf(`</%s>`, tag)
}

func Div(vars *map[string]string, pars ...string) (out string) {
	return getTag(`div`, pars...)
}

func P(vars *map[string]string, pars ...string) (out string) {
	return getTag(`p`, pars...)
}

func Em(vars *map[string]string, pars ...string) (out string) {
	return getTag(`em`, pars...)
}

func Small(vars *map[string]string, pars ...string) (out string) {
	return getTag(`small`, pars...)
}

func Divs(vars *map[string]string, pars ...string) (out string) {
	count := 0
	for _, item := range pars {
		out += fmt.Sprintf(`<div class="%s">`, getClass(item))
		count++
	}
	if val, ok := (*vars)[`divs`]; ok {
		(*vars)[`divs`] = fmt.Sprintf(`%s,%d`, val, count)
	} else {
		(*vars)[`divs`] = fmt.Sprintf(`%d`, count)
	}
	return
}

func DivsEnd(vars *map[string]string, pars ...string) (out string) {
	if val, ok := (*vars)[`divs`]; ok && len(val) > 0 {
		divs := strings.Split(val, `,`)
		out = strings.Repeat(`</div>`, StrToInt(divs[len(divs)-1]))
		(*vars)[`divs`] = strings.Join(divs[:len(divs)-1], `,`)
	}
	return
}

func SetVar(vars *map[string]string, pars ...string) string {
	for _, item := range pars {
		var proc bool
		var val string
		lr := strings.SplitN(item, `#=`, 2)
		if len(lr) != 2 {
			lr = strings.SplitN(item, `=`, 2)
			if len(lr) != 2 {
				continue
			}
			proc = true
		}
		if proc {
			val = textproc.Process(lr[1], vars)
			if val == `NULL` {
				val = textproc.Macro(lr[1], vars)
			}
		} else {
			val = lr[1]
		}
		val = strings.Replace(val, `#!`, `#`, -1)
		(*vars)[strings.TrimSpace(lr[0])] = strings.Trim(val, " `\"")
	}
	return ``
}

func TextHidden(vars *map[string]string, pars ...string) (out string) {
	for _, item := range pars {
		out += fmt.Sprintf(`<textarea style="display:none;" id="%s">%s</textarea>`, item, (*vars)[item])
	}
	return
}

func TxId(vars *map[string]string, pars ...string) string {
	if len(pars) == 0 {
		return `0`
	}
	return Int64ToStr(TypeInt(pars[0]))
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
	if params == `''` {
		params = ``
	}
	class := `"btn btn-primary"`
	if len(pars) >= 4 {
		class = pars[3]
	}
	return fmt.Sprintf(`<button type="button" class=%s onclick="load_template('%s', {%s} )">%s</button>`, class, pars[0], params, pars[1])
}

func BtnSys(vars *map[string]string, pars ...string) string {
	params := ``
	if len(pars) < 2 {
		return ``
	}
	if len(pars) >= 3 {
		params = pars[2]
	}
	if params == `''` {
		params = ``
	}
	class := `"btn btn-primary"`
	if len(pars) >= 4 {
		class = pars[3]
	}
	return fmt.Sprintf(`<button type="button" class=%s onclick="load_page('%s', {%s} )">%s</button>`, class, pars[0], params, pars[1])
}

func StateLink(vars *map[string]string, pars ...string) string {
	if len(pars) < 2 {
		return ``
	}
	return (*vars)[fmt.Sprintf(`%s_%s`, pars[0], pars[1])]
}

func Table(vars *map[string]string, pars *map[string]string) string {
	fields := `*`
	order := ``
	where := ``
	limit := ``
	if val, ok := (*pars)[`Order`]; ok {
		order = `order by ` + lib.Escape(val)
	}
	if val, ok := (*pars)[`Where`]; ok {
		where = `where ` + lib.Escape(val)
	}
	if val, ok := (*pars)[`Limit`]; ok && len(val) > 0 {
		opar := strings.Split(val, `,`)
		if len(opar) == 1 {
			limit = fmt.Sprintf(` limit %d`, StrToInt64(opar[0]))
		} else {
			limit = fmt.Sprintf(` offset %d limit %d`, StrToInt64(opar[0]), StrToInt64(opar[1]))
		}
	}
	if val, ok := (*pars)[`Fields`]; ok {
		fields = lib.Escape(val)
	}
	list, err := DB.GetAll(fmt.Sprintf(`select %s from %s %s %s%s`, fields,
		lib.EscapeName((*pars)[`Table`]), where, order, limit), -1)
	if err != nil {
		return err.Error()
	}
	columns := textproc.Split((*pars)[`Columns`])
	out := `<table  class="table table-striped table-bordered table-hover"><tr>`
	for _, th := range *columns {
		out += `<th>` + th[0] + `</th>`
		th[1] = strings.TrimSpace(th[1])
		if strings.HasPrefix(th[1], `StateLink`) && strings.IndexByte(th[1], ',') > 0 {
			linklist := strings.TrimSpace(th[1][strings.IndexByte(th[1], '(')+1 : strings.IndexByte(th[1], ',')])
			if alist := strings.Split(StateValue(vars, linklist), `,`); len(alist) > 0 {
				for ind, item := range alist {
					(*vars)[fmt.Sprintf(`%s_%d`, linklist, ind+1)] = LangText(item, int(StrToInt64((*vars)[`state_id`])), (*vars)[`accept_lang`])
				}
			}
		}
	}
	out += `</tr>`
	for _, item := range list {
		out += `<tr>`
		for key, value := range item {
			if key != `state_id` {
				(*vars)[key] = value
			}
		}
		for _, th := range *columns {
			//			val := textproc.Process(th[1], vars)
			//			if val == `NULL` {
			val := textproc.Macro(th[1], vars)
			//			}
			out += `<td>` + strings.Replace(val, "\n", "\n<br>", -1) + `</td>`
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
	class := ``
	if len(pars) > 1 {
		alt = pars[1]
	}
	if len(pars) > 2 {
		class = pars[2]
	}
	rez := " "
	if len(pars[0]) > 0 && (strings.HasPrefix(pars[0], `data:`) || strings.HasSuffix(pars[0], `jpg`) ||
		strings.HasSuffix(pars[0], `png`)) {
		rez = fmt.Sprintf(`<img src="%s" class="%s" alt="%s" stylex="display:block;">`, pars[0], class, alt)
	}
	return rez
}

func StateValue(vars *map[string]string, pars ...string) string {
	val, _ := StateParam(StrToInt64((*vars)[`state_id`]), pars[0])
	return val
}

func LiTemplate(vars *map[string]string, pars ...string) string {
	name := pars[0]
	title := name
	params := ``
	if len(pars) > 1 {
		title = pars[1]
	}
	if len(pars) >= 3 {
		params = pars[2]
	}
	return fmt.Sprintf(`<li><a href="#" onclick="load_template('%s', {%s}); HideMenu();"><span>%s</span></a></li>`,
		name, params, title)
}

func AppNav(vars *map[string]string, pars ...string) string {
	name := pars[0]
	title := name
	if len(pars) > 1 {
		title = pars[1]
	}
	return fmt.Sprintf(`<a href="#" onclick="load_app('%s'); HideMenu();"><span>%s</span></a>`, name, title)
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
	<a href='#' onclick="load_page('editPage', {name: '#page#', global:'#global#'} )">Edit</a></span>%s</ol>`,
		strings.Join(li, `&nbsp;/&nbsp;`)), vars)
}

func SysLink(vars *map[string]string, pars ...string) string {
	params := ``
	if len(pars) < 2 {
		return ``
	}
	if len(pars) >= 3 {
		params = pars[2]
	}
	return fmt.Sprintf(`<a href='#'onclick="load_page('%s', {%s} )">%s</a>`, pars[0], params, pars[1])
}

func MarkDown(vars *map[string]string, pars ...string) string {
	return textproc.Macro(string(blackfriday.MarkdownCommon([]byte(pars[0]))), vars)
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

func Form(vars *map[string]string, pars ...string) string {
	var class string
	if len(pars[0]) > 0 {
		class = fmt.Sprintf(`class="%s"`, pars[0])
	}
	return fmt.Sprintf(`<form role="form" %s>`, class)
}

func FormEnd(vars *map[string]string, pars ...string) string {
	return `</form>`
}

func Label(vars *map[string]string, pars ...string) string {
	var class string
	if len(pars) > 1 && len(pars[1]) > 0 {
		class = fmt.Sprintf(`class="%s"`, pars[1])
	}
	text := LangText(pars[0], int(StrToInt64((*vars)[`state_id`])), (*vars)[`accept_lang`])
	return fmt.Sprintf(`<label %s>%s</label>`, class, text)
}

func ValueById(vars *map[string]string, pars ...string) string {
	// tablename, id of value, parameters
	if len(pars) < 3 {
		return ``
	}
	value, err := DB.OneRow(`select * from ` + lib.EscapeName(pars[0]) + ` where id='` + lib.Escape(pars[1]) + `'`).String()
	if err != nil {
		return err.Error()
	}
	keys := make(map[string]string)
	src := strings.Split(lib.Escape(pars[2]), `,`)
	if len(pars) == 4 {
		dest := strings.Split(lib.Escape(pars[3]), `,`)
		for i, val := range src {
			if len(dest) > i {
				keys[val] = dest[i]
			}
		}
	}
	if len(value) > 0 {
		for _, key := range src {
			val := value[key]
			if val == `NULL` {
				val = ``
			}
			if ikey, ok := keys[key]; ok {
				(*vars)[ikey] = val
			} else {
				(*vars)[key] = val
			}
		}
	}
	return ``
}

func TXButton(vars *map[string]string, pars *map[string]string) string {
	var unique int64
	if uval, ok := (*vars)[`tx_unique`]; ok {
		unique = StrToInt64(uval) + 1
	}
	(*vars)[`tx_unique`] = Int64ToStr(unique)
	name := (*pars)[`Contract`]
	//	init := (*pars)[`Init`]
	fmt.Println(`TXButton Init`, *vars)
	onsuccess := (*pars)[`OnSuccess`]
	contract := smart.GetContract(name, uint32(StrToUint64((*vars)[`state_id`])))
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
	data, err := static.Asset("static/tx_button.html")

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

	if len(onsuccess) > 0 {
		pars := strings.SplitN(onsuccess, `,`, 3)
		onsuccess = ``
		if len(pars) >= 2 {
			onsuccess = fmt.Sprintf(`load_%s('%s'`, pars[0], pars[1])
			if len(pars) == 3 {
				onsuccess += `,{` + pars[2] + `}`
			}

			onsuccess += `)`
		} else {
			onsuccess = lib.Escape(pars[0])
		}
	}

	b := new(bytes.Buffer)
	finfo := TxButtonInfo{TxName: name, Unique: template.JS((*vars)[`tx_unique`]), OnSuccess: template.JS(onsuccess),
		Fields: make([]TxInfo, 0), Data: FormCommon{
			CountSignArr: []byte{1}}}

	idnames := strings.Split((*pars)[`Inputs`], `,`)
	names := make(map[string]string)
	for _, idn := range idnames {
		lr := strings.SplitN(idn, `=`, 2)
		if len(lr) == 2 {
			names[lr[0]] = lr[1]
		}
	}
txlist:
	for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
		idname := fitem.Name
		if idn, ok := names[idname]; ok {
			idname = idn
		}
		for _, tag := range []string{`date`, `polymap`, `map`, `image`, `text`, `address`} {
			if strings.Index(fitem.Tags, tag) >= 0 {
				finfo.Fields = append(finfo.Fields, TxInfo{Name: fitem.Name, Id: idname, HtmlType: tag})
				continue txlist
			}
		}
		finfo.Fields = append(finfo.Fields, TxInfo{Name: fitem.Name, Id: idname, HtmlType: "textinput"})
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

func getSelect(linklist string) (data []map[string]string, id string, name string, err error) {
	var count int64
	tbl := strings.Split(linklist, `.`)
	tblname := lib.EscapeName(tbl[0])
	name = tbl[1]
	id = `id`
	if len(tbl) > 2 {
		id = tbl[2]
	}
	count, err = DB.Single(`select count(*) from ` + tblname).Int64()
	if err != nil {
		return
	}
	if count > 0 && count <= 50 {
		data, err = DB.GetAll(fmt.Sprintf(`select %s, %s from %s order by %s`, id,
			lib.EscapeName(name), tblname, lib.EscapeName(name)), -1)
	}
	return
}

func TXForm(vars *map[string]string, pars *map[string]string) string {
	var unique int64
	if uval, ok := (*vars)[`tx_unique`]; ok {
		unique = StrToInt64(uval) + 1
	}
	(*vars)[`tx_unique`] = Int64ToStr(unique)
	name := (*pars)[`Contract`]
	//	init := (*pars)[`Init`]
	//fmt.Println(`TXForm Init`, *vars)
	onsuccess := (*pars)[`OnSuccess`]
	contract := smart.GetContract(name, uint32(StrToUint64((*vars)[`state_id`])))
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

	if len(onsuccess) > 0 {
		pars := strings.SplitN(onsuccess, `,`, 3)
		onsuccess = ``
		if len(pars) >= 2 {
			onsuccess = fmt.Sprintf(`load_%s('%s'`, pars[0], pars[1])
			if len(pars) == 3 {
				onsuccess += `,{` + pars[2] + `}`
			}

			onsuccess += `)`
		} else {
			onsuccess = lib.Escape(pars[0])
		}
	}

	b := new(bytes.Buffer)
	finfo := FormInfo{TxName: name, Unique: template.JS((*vars)[`tx_unique`]), OnSuccess: template.JS(onsuccess), Fields: make([]FieldInfo, 0), Data: FormCommon{
		CountSignArr: []byte{1}}}

	gettag := func(prefix uint8, def, tags string) string {
		ret := def
		if off := strings.IndexByte(tags, prefix); off >= 0 {
			end := off + 1
			for end < len(tags) {
				if tags[end] == ' ' {
					break
				}
				end++
			}
			ret = tags[off+1 : end]
		}
		return ret
	}
	getlang := func(res string) string {
		return LangText(res, int(StrToInt64((*vars)[`state_id`])), (*vars)[`accept_lang`])
	}
txlist:
	for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
		var value string
		if val, ok := (*vars)[fitem.Name]; ok {
			value = val
		}
		if strings.Index(fitem.Tags, `hidden`) >= 0 {
			finfo.Fields = append(finfo.Fields, FieldInfo{Name: fitem.Name, HtmlType: `hidden`,
				TxType: fitem.Type.String(), Title: ``, Value: value})
			continue
		}
		langres := gettag('#', fitem.Name, fitem.Tags)
		linklist := gettag('@', ``, fitem.Tags)
		title := getlang(langres)
		for _, tag := range []string{`date`, `polymap`, `map`, `image`, `text`, `address`} {
			if strings.Index(fitem.Tags, tag) >= 0 {
				finfo.Fields = append(finfo.Fields, FieldInfo{Name: fitem.Name, HtmlType: tag,
					TxType: fitem.Type.String(), Title: title, Value: value})
				continue txlist
			}
		}
		if len(linklist) > 0 {
			sellist := SelList{StrToInt64(value), make(map[int]string)}
			if strings.IndexByte(linklist, '.') >= 0 {
				if data, id, name, err := getSelect(linklist); err != nil {
					return err.Error()
				} else if len(data) > 0 {
					for _, item := range data {
						sellist.List[int(StrToInt64(item[id]))] = item[name]
					}
				}
			} else if alist := strings.Split(StateValue(vars, linklist), `,`); len(alist) > 0 {
				for ind, item := range alist {
					sellist.List[ind+1] = getlang(item)
				}
			}
			finfo.Fields = append(finfo.Fields, FieldInfo{Name: fitem.Name, HtmlType: "select",
				TxType: fitem.Type.String(), Title: title, Value: sellist})
		} else if fitem.Type.String() == `string` || fitem.Type.String() == `int64` || fitem.Type.String() == `float64` ||
			fitem.Type.String() == `decimal.Decimal` {
			finfo.Fields = append(finfo.Fields, FieldInfo{Name: fitem.Name, HtmlType: "textinput",
				TxType: fitem.Type.String(), Title: title, Value: value})
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

func IdToAddress(vars *map[string]string, pars ...string) string {
	var idval string
	if len(pars) == 0 || len(pars[0]) == 0 {
		idval = (*vars)[`citizen`]
	} else {
		idval = pars[0]
	}
	id, _ := strconv.ParseInt(idval, 10, 64)
	if id == 0 {
		return `unknown address`
	}
	return lib.AddressToString(uint64(id))
}

func Ring(vars *map[string]string, pars ...string) string {
	class := `col-md-4`
	title := ``
	count := ``
	size := 18
	if len(pars) > 2 {
		size = int(StrToInt64(pars[2]))
	}
	if len(pars) > 1 {
		title = getClass(pars[1])
	}
	if len(pars) > 0 {
		count = lib.NumString(pars[0])
	}
	return fmt.Sprintf(`<div class="%s"><div class="panel panel-default"> <div class="panel-body">
			<div class="text-info">%s</div>
			<div class="population" style="font-size:%dpx"><img src="static/img/spacer.gif" alt=""><span>%s</span></div>
		 </div></div></div>`, class, title, size, count)
}

func WiBalance(vars *map[string]string, pars ...string) string {
	if len(pars) != 2 {
		return ``
	}
	return fmt.Sprintf(`<div class="panel widget"><div class="row row-table row-flush">
			<div class="col-xs-4 bg-info text-center"><em class="glyphicons glyphicons-coins x2"></em>
			</div><div class="col-xs-8">
			   <div class="panel-body text-center">
				  <h4 class="mt0">%s %s</h4>
				  <p class="mb0 text-muted">Balance</p>
			   </div></div></div></div>`, lib.NumString(pars[0]), lib.Escape(pars[1]))
}

func WiAccount(vars *map[string]string, pars ...string) string {
	if len(pars) != 1 {
		return ``
	}
	return fmt.Sprintf(`<div class="panel widget bg-success"><div class="row row-table">
			<div class="col-xs-4 text-center bg-success-dark pv-lg">
			   <em class="glyphicons glyphicons-credit-card x2"></em></div>
			<div class="col-xs-8 pv-lg">
			   <div class="h1 m0 text-bold">%s</div>
			   <div class="text-uppercase">ACCOUNT NUMBER</div>
			</div></div></div>`, lib.Escape(pars[0]))
}

func WiCitizen(vars *map[string]string, pars ...string) string {
	image := `/static/img/apps/ava.png`
	flag := ``
	if len(pars) < 2 {
		return ``
	}
	if len(pars) > 2 && pars[2] != `NULL` && pars[2] != `#my_avatar#` {
		image = pars[2]
	}
	if len(pars) > 3 {
		flag = fmt.Sprintf(`<img src="%s" alt="Image" class="wd-xs">`, pars[3])
	}
	address := lib.AddressToString(uint64(StrToInt64(pars[1])))
	(*vars)["wicitizen"] = `1`
	return fmt.Sprintf(`<div class="panel widget"><div class="panel-body">
			<div class="row row-table"><div class="col-xs-6 text-center">
				  <img src="%s" alt="Image" class="img-circle thumb96">
			   </div>
			   <div class="col-xs-6">
				  <h3 class="mt0">%s</h3>
				  <ul class="list-unstyled">
					 <li class="mb-sm">
					 	%s
					 </li></ul></div></div></div>
		 <div class="panel-body bg-inverse"><div class="row row-table text-center">
			   <div class="col-xs-12 p0">
				  <p class="m0 h4">%s <i class="clipboard fa fa-clipboard" aria-hidden="true" data-clipboard-action="copy" 
				  data-clipboard-text="%s" onClick="CopyToClipboard('.clipboard')"  data-notify="" 
				  data-message="Copied to clipboard" data-options="{&quot;status&quot;:&quot;info&quot;}"></i></p>
				  <p class="m0 text-muted">Citizen ID</p>
		</div></div></div></div>`, image, lib.Escape(pars[0]), flag, address, address)
}

func Mult(vars *map[string]string, pars ...string) string {
	if len(pars) != 2 {
		return ``
	}
	return Int64ToStr(round(StrToFloat64(pars[0]) * StrToFloat64(pars[1])))
}

func Date(vars *map[string]string, pars ...string) string {
	if len(pars) == 0 || pars[0] == `NULL` {
		return ``
	}
	itime, err := time.Parse(`2006-01-02T15:04:05`, pars[0][:19])
	if err != nil {
		return err.Error()
	}
	var format string
	if len(pars) > 1 {
		format = pars[1]
	} else {
		format = LangText(`dateformat`, int(StrToInt64((*vars)[`state_id`])), (*vars)[`accept_lang`])
		if format == `dateformat` {
			format = `2006-01-02`
		}
	}
	format = strings.Replace(format, `YYYY`, `2006`, -1)
	format = strings.Replace(format, `YY`, `06`, -1)
	format = strings.Replace(format, `MM`, `01`, -1)
	format = strings.Replace(format, `DD`, `02`, -1)

	return itime.Format(format)
}

func DateTime(vars *map[string]string, pars ...string) string {
	if len(pars) == 0 || pars[0] == `NULL` {
		return ``
	}
	itime, err := time.Parse(`2006-01-02T15:04:05`, pars[0][:19])
	if err != nil {
		return err.Error()
	}
	var format string
	if len(pars) > 1 {
		format = pars[1]
	} else {
		format = LangText(`timeformat`, int(StrToInt64((*vars)[`state_id`])), (*vars)[`accept_lang`])
		if format == `timeformat` {
			format = `2006-01-02 15:04:05`
		}
	}
	format = strings.Replace(format, `YYYY`, `2006`, -1)
	format = strings.Replace(format, `YY`, `06`, -1)
	format = strings.Replace(format, `MM`, `01`, -1)
	format = strings.Replace(format, `DD`, `02`, -1)
	format = strings.Replace(format, `HH`, `15`, -1)
	format = strings.Replace(format, `MI`, `04`, -1)
	format = strings.Replace(format, `SS`, `05`, -1)

	return itime.Format(format)
}

func Select(vars *map[string]string, pars ...string) string {
	var (
		class string
		value int64
	)
	list := make([]SelInfo, 0)
	if len(pars) > 1 {
		if strings.IndexByte(pars[1], '.') >= 0 {
			if data, id, name, err := getSelect(pars[1]); err != nil {
				return err.Error()
			} else if len(data) > 0 {
				for _, item := range data {
					list = append(list, SelInfo{Id: StrToInt64(item[id]), Name: item[name]})
				}
			}
		} else if alist := strings.Split(StateValue(vars, pars[1]), `,`); len(alist) > 0 {
			for ind, item := range alist {
				list = append(list, SelInfo{Id: int64(ind + 1), Name: LangText(item, int(StrToInt64((*vars)[`state_id`])), (*vars)[`accept_lang`])})
			}
		}
	}
	if len(pars) > 2 {
		class = pars[2]
	}
	if len(pars) > 3 {
		value = StrToInt64(pars[3])
	}

	out := fmt.Sprintf(`<select id="%s" class="selectbox form-control %s">`, pars[0], class)
	for _, item := range list {
		var selected string
		if item.Id == value {
			selected = `selected`
		}
		out += fmt.Sprintf(`<option value="%d" %s>%s</option>`, item.Id, selected, item.Name)

	}
	return out + `</select>`
}

func Map(vars *map[string]string, pars ...string) string {
	(*vars)[`wimap`] = `1`
	return fmt.Sprintf(`<div class="wimap">%s</div>`, pars[0])
}

func MapPoint(vars *map[string]string, pars ...string) string {
	(*vars)[`wimappoint`] = `1`
	return fmt.Sprintf(`<div class="wimappoint">%s</div>`, pars[0])
}

/*func AddressToId(vars *map[string]string, pars ...string) string {
	var idval int64
	if len(pars) == 0 || len(pars[0]) == 0 {
		uid,_ := strconv.ParseInt((*vars)[`citizen`], 10, 64)
		idval = int64(uid)
	} else {
		if len(pars[0]) > 21 {
			idval = lib.StringToAddress(pars[0])
		} else {
			if pars[0][0] == '-' {
				idval,_ = strconv.ParseInt(pars[0], 10, 64)
			} else {
				uid,_ := strconv.ParseUint(pars[0], 10, 64)
				idval = int64(uid)
			}
		}
	}
	return fmt.Sprintf(`%d`, idval)
}*/

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
	//	fmt.Println(`PROC`, err)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
