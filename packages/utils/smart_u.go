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
	"regexp"
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
	Param    string      `json:"param"`
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
	AutoClose bool
	Silent    bool
	Data      FormCommon
}

type TxInfo struct {
	Name     string `json:"name"`
	Id       string `json:"id"`
	Value    string `json:"value"`
	HtmlType string `json:"htmlType"`
	Param    string `json:"param"`
}

type TxButtonInfo struct {
	TxName    string
	Name      string
	Class     string
	ClassBtn  string
	Unique    template.JS
	OnSuccess template.JS
	Fields    []TxInfo
	AutoClose bool
	Silent    bool
	Data      FormCommon
}

type TxBtnCont struct {
	TxName    string
	Name      string
	Class     string
	ClassBtn  string
	Unique    template.JS
	OnSuccess template.JS
	Fields    []TxInfo
	AutoClose bool
	Silent    bool
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

	textproc.AddMaps(&map[string]textproc.MapFunc{`Table`: Table, `TxForm`: TxForm, `TxButton`: TXButton,
		`ChartPie`: ChartPie, `ChartBar`: ChartBar})
	textproc.AddFuncs(&map[string]textproc.TextFunc{`Address`: IdToAddress, `BtnEdit`: BtnEdit,
		`Image`: Image, `Div`: Div, `P`: Par, `Em`: Em, `Small`: Small, `A`: A, `Span`: Span, `Strong`: Strong, `Divs`: Divs, `DivsEnd`: DivsEnd,
		`LiTemplate`: LiTemplate, `LinkTemplate`: LinkTemplate, `BtnPage`: BtnPage,
		`CmpTime`: CmpTime, `Title`: Title, `MarkDown`: MarkDown, `Navigation`: Navigation, `PageTitle`: PageTitle,
		`PageEnd`: PageEnd, `StateValue`: StateValue, `Json`: JsonScript, `And`: And, `Or`: Or,
		`TxId`: TxId, `SetVar`: SetVar, `GetList`: GetList, `GetRow`: GetRowVars, `GetOne`: GetOne, `TextHidden`: TextHidden,
		`ValueById`: ValueById, `FullScreen`: FullScreen, `Ring`: Ring, `WiBalance`: WiBalance,
		`WiAccount`: WiAccount, `WiCitizen`: WiCitizen, `Map`: Map, `MapPoint`: MapPoint, `StateLink`: StateLink,
		`If`: If, `IfEnd`: IfEnd, `Else`: Else, `ElseIf`: ElseIf, `Trim`: Trim, `Date`: Date, `DateTime`: DateTime, `Now`: Now, `Input`: Input,
		`Textarea`: Textarea, `InputMoney`: InputMoney, `InputAddress`: InputAddress, `ForList`: ForList, `ForListEnd`: ForListEnd,
		`BlockInfo`: BlockInfo, `Back`: Back, `ListVal`: ListVal, `Tag`: Tag, `BtnContract`: BtnContract,
		`Form`: Form, `FormEnd`: FormEnd, `Label`: Label, `Legend`: Legend, `Select`: Select, `Param`: Param, `Mult`: Mult,
		`Money`: Money, `Source`: Source, `Val`: Val, `Lang`: LangRes, `LangJS`: LangJS, `InputDate`: InputDate,
		`MenuGroup`: MenuGroup, `MenuEnd`: MenuEnd, `MenuItem`: MenuItem, `MenuPage`: MenuPage, `MenuBack`: MenuBack, `WhiteMobileBg`: WhiteMobileBg, `Bin2Hex`: Bin2Hex, `MessageBoard`: MessageBoard,
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
			break
		}
	}
	smart.ExternOff()
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
			log.Error("Load Contract", item[`name`], err)
			fmt.Println("Error Load Contract", item[`name`], err)
			//return
		} else {
			fmt.Println("OK Load Contract", item[`name`])
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

func EGSRate(idstate int64) (float64, error) {
	return DB.Single(`SELECT value FROM "`+Int64ToStr(idstate)+`_state_parameters" WHERE name = ?`, `egs_rate`).Float64()
}

func StateParam(idstate int64, name string) (string, error) {
	return DB.Single(`SELECT value FROM "`+Int64ToStr(idstate)+`_state_parameters" WHERE name = ?`, name).String()
}

func Param(vars *map[string]string, pars ...string) string {
	if val, ok := (*vars)[pars[0]]; ok {
		return val
	}
	return ``
}

func LangRes(vars *map[string]string, pars ...string) string {
	ret, _ := LangText(pars[0], int(StrToInt64((*vars)[`state_id`])), (*vars)[`accept_lang`])
	return ret
}

func LangJS(vars *map[string]string, pars ...string) string {
	return fmt.Sprintf(`<span class="lang" lang-id="%s"></span>`, pars[0])
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
		ret0, _ := decimal.NewFromString(strings.TrimSpace(cond[0]))
		ret1, _ := decimal.NewFromString(strings.TrimSpace(cond[1]))
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

func Money(vars *map[string]string, pars ...string) string {
	var cents int
	if len(pars) > 1 {
		cents = StrToInt(pars[1])
	} else {
		cents = StrToInt(StateValue(vars, `money_digit`))
	}
	ret := pars[0]
	if cents > 0 && strings.IndexByte(ret, '.') < 0 {
		if len(ret) < cents+1 {
			ret = strings.Repeat(`0`, cents+1-len(ret)) + ret
		}
		ret = ret[:len(ret)-cents] + `.` + ret[len(ret)-cents:]
	}
	return ret
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
	if len(pars) == 1 && strings.HasSuffix((*vars)[`ifs`], `0`) {
		(*vars)[`ifs`] = (*vars)[`ifs`] + `0`
		return ``
	}
	isTrue := ifValue(pars[0])
	if len(pars) == 1 {
		state := `0`
		if isTrue {
			state = `1`
		}
		(*vars)[`ifs`] = (*vars)[`ifs`] + state
		return ``
	}
	if isTrue {
		return pars[1]
	}
	if len(pars) > 2 {
		return pars[2]
	}
	return ``
}

func Else(vars *map[string]string, pars ...string) string {
	ival := []byte((*vars)[`ifs`])
	if ilen := len(ival); ilen == 1 || (ilen > 1 && ival[ilen-2] == '1') {
		if ival[ilen-1] == '0' {
			ival[ilen-1] = '1'
		} else {
			ival[ilen-1] = '0'
		}
		(*vars)[`ifs`] = string(ival)
	}
	return ``
}

func ElseIf(vars *map[string]string, pars ...string) string {
	ival := []byte((*vars)[`ifs`])
	if ilen := len(ival); ilen == 1 || (ilen > 1 && ival[ilen-2] == '1') {
		if ival[ilen-1] == '0' {
			if ifValue(pars[0]) {
				ival[ilen-1] = '1'
			} else {
				ival[ilen-1] = '0'
			}
		} else {
			ival[ilen-1] = '-'
		}
		(*vars)[`ifs`] = string(ival)
	}
	return ``
}

func IfEnd(vars *map[string]string, pars ...string) string {
	ilen := len((*vars)[`ifs`])
	if ilen > 0 {
		(*vars)[`ifs`] = (*vars)[`ifs`][:ilen-1]
	}
	return ``
}

func Now(vars *map[string]string, pars ...string) string {
	var (
		cut             int
		query, interval string
	)
	if len(pars) > 1 && len(pars[1]) > 0 {
		interval = pars[1]
		if interval[0] != '-' && interval[0] != '+' {
			interval = `+` + interval
		}
		interval = fmt.Sprintf(` %s interval '%s'`, interval[:1], strings.TrimSpace(interval[1:]))
	}
	if pars[0] == `` {
		query = `select round(extract(epoch from now()` + interval + `))::integer`
		cut = 10
	} else {
		query = `select now()` + interval
		switch pars[0] {
		case `datetime`:
			cut = 19
		default:
			format := pars[0]
			if strings.Index(format, `HH`) >= 0 && strings.Index(format, `HH24`) < 0 {
				format = strings.Replace(format, `HH`, `HH24`, -1)
			}
			query = fmt.Sprintf(`select to_char(now()%s, '%s')`, interval, format)
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

func Textarea(vars *map[string]string, pars ...string) string {
	var (
		class, value string
	)
	if len(pars) > 1 {
		class = pars[1]
	}
	if len(pars) > 2 {
		value = pars[2]
	}
	return fmt.Sprintf(`<textarea id="%s" class="%s">%s</textarea>`,
		pars[0], class, value)
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
		placeholder = LangRes(vars, pars[2])
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

func InputDate(vars *map[string]string, pars ...string) string {
	var (
		class, value string
	)
	if len(pars) > 1 {
		class = pars[1]
	}
	if len(pars) > 2 {
		value = pars[2]
	}
	(*vars)["widate"] = `1`
	return fmt.Sprintf(`<input type="text" class="datetimepicker %s" id="%s" value="%s">`, class, pars[0], value)
}

func InputMoney(vars *map[string]string, pars ...string) string {
	var (
		class, value string
		digit        int
	)
	if len(pars) > 1 {
		class = pars[1]
	}
	if len(pars) > 3 {
		digit = StrToInt(pars[3])
	} else {
		digit = StrToInt(StateValue(vars, `money_digit`))
	}
	if len(pars) > 2 {
		value = Money(vars, pars[2], IntToStr(digit))
	}
	(*vars)["wimoney"] = `1`
	return fmt.Sprintf(`<input id="%s" type="text" value="%s"
				data-inputmask="'alias': 'numeric', 'rightAlign': false, 'groupSeparator': ' ', 'autoGroup': true, 'digits': %d, 'digitsOptional': false, 'prefix': '', 'placeholder': '0'"
	class="inputmask %s">`, pars[0], value, digit, class)
}

func InputAddress(vars *map[string]string, pars ...string) string {
	var (
		class, value string
	)
	if len(pars) > 1 {
		class = pars[1]
	}
	if len(pars) > 2 {
		value = pars[2]
	}
	(*vars)["wiaddress"] = `1`
	return fmt.Sprintf(`<input id="%s" type="text" value="%s" data-type="wallet" class="%s address">
				<ul class="parsley-errors-list">
					<li class="parsley-required">Please enter the correct address</li>
				</ul>`, pars[0], value, class)
}

func Trim(vars *map[string]string, pars ...string) string {
	if len(pars) == 0 {
		return ``
	}
	return strings.TrimSpace(pars[0])
}

func Back(vars *map[string]string, pars ...string) string {
	if len(pars[0]) == 0 || len(pars) < 2 || len(pars[1]) == 0 {
		return ``
	}
	var params string
	if len(pars) == 3 {
		params = lib.Escape(pars[2])
	}
	return fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
	hist_push(['load_%s', '%s', {%s}]);
</script>`, lib.Escape(pars[0]), lib.Escape(pars[1]), params)
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

func WhiteMobileBg(vars *map[string]string, pars ...string) string {
	wide := `add`
	if len(pars) > 0 && pars[0] == `0` {
		wide = `remove`
	}
	return fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
	$("body").%sClass('flatPageMobile');
</script>`, wide)
}

func Bin2Hex(vars *map[string]string, pars ...string) string {
	if len(pars) == 0 {
		return ``
	}
	return string(BinToHex(pars[0]))
}

func WhiteBg(vars *map[string]string, pars ...string) string {
	wide := `add`
	if len(pars) > 0 && pars[0] == `0` {
		wide = `remove`
	}
	return fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
	$("body").%sClass('flatPage');
</script>`, wide)
}

func MessageBoard(vars *map[string]string, pars ...string) string {
	messages, err := DB.GetAll(`select * from "global_messages" order by id`, 100)
	if err != nil {
		return ``
	}
	ret := ``
	for _, item := range messages {
		ret += `<a href="#" class="list-group-item">
						<div class="media-box">
							<div class="pull-left">
								<img src="` + item[`ava`] + `" alt="Image" class="media-box-object img-circle thumb32">
							</div>
							<div class="media-box-body clearfix">
								<small class="flag ru pull-right">` + item[`flag`] + `</small>
								<strong class="media-box-heading text-primary">` + item[`username`] + `</strong>
								<p class="mb-sm pr-lg">
									<small>` + item[`text`] + `</small>
								</p>
							</div>
						</div>
					</a>`

	}
	//	(*vars)["wibtncont"] = `1`
	return fmt.Sprintf(`<div id="panelDemo2" class="panel panel-info elastic" data-sweet-alert>
			<div class="panel-heading">
				UN Conference
				<div data-widget="panel-collapse"></div>
			</div>
			<div class="panel-body">
				<div data-widget="panel-scroll">
					%s
				</div>
			</div>
			<div class="panel-footer"><div class="input-group">
                                 <input placeholder="press message" class="form-control input-sm" type="text" id="message_board_text" value="Hello">
                                 <span class="input-group-btn">`+
		TXButton(vars, &map[string]string{`Contract`: `addMessage`, `Name`: `Send`,
			`ClassBtn`: `btn btn-default btn-sm`, `AutoClose`: `1`,
			`OnSuccess`: `template,dashboard_default`, `Inputs`: `Text=message_board_text`})+
		/*                                <script>
										 function send_mess(obj) {
		                                 	var message_board_text =  $( "#message_board_text" ).val();
		  								 	btn_contract(obj, 'addMessage', {Text: message_board_text}, 'You vote for candidate to #campaign#',
											        'template', 'dashboard_default', {})
										 }
		                                 </script>
		                                    <button type="button" class="btn btn-default btn-sm" data-tool="panel-refresh" onclick="send_mess(this)" id="panelRefresh_1">Send</button>
		                                    </button>*/
		`                                 </span>
                              </div></div>
		</div>`, ret)
}

func GetList(vars *map[string]string, pars ...string) string {
	// name, table, fields, where, order, limit
	if len(pars) < 3 {
		return ``
	}
	where := ``
	order := ``
	limit := -1
	fields := lib.Escape(pars[2])
	keys := strings.Split(fields, `,`)
	if len(pars) >= 4 {
		where = ` where ` + lib.Escape(pars[3])
	}
	if len(pars) >= 5 {
		order = ` order by ` + lib.EscapeName(pars[4])
	}
	if len(pars) >= 6 {
		limit = StrToInt(pars[5])
	}

	value, err := DB.GetAll(`select `+fields+` from `+lib.EscapeName(pars[1])+where+order, limit)
	if err != nil {
		return err.Error()
	}
	list := make([]string, 0)
	cols := make([]string, 0)
	//out := make(map[string]map[string]string)
	for _, item := range value {
		ikey := item[keys[0]]
		list = append(list, ikey)
		if len(cols) == 0 {
			for key := range item {
				cols = append(cols, key)
			}
		}
		for key, ival := range item {
			if strings.IndexByte(ival, '<') >= 0 {
				//				item[key] = lib.StripTags(ival)
				ival = lib.StripTags(ival)
			}
			(*vars)[pars[0]+ikey+key] = ival
		}
		//		out[item[keys[0]]] = item
	}
	(*vars)[pars[0]+`_list`] = strings.Join(list, `|`)
	(*vars)[pars[0]+`_columns`] = strings.Join(cols, `|`)
	return ``
}

func ForList(vars *map[string]string, pars ...string) string {
	(*vars)[`for_name`] = pars[0]
	(*vars)[`for_loop`] = `1`
	return ``
}

func ForListEnd(vars *map[string]string, pars ...string) (out string) {
	name := (*vars)[`for_name`]
	list := strings.Split((*vars)[name+`_list`], `|`)
	cols := strings.Split((*vars)[name+`_columns`], `|`)
	for i, item := range list {
		item = strings.TrimSpace(item)
		if len(item) == 0 {
			continue
		}
		(*vars)[`index`] = fmt.Sprintf(`%d`, i+1)
		for _, icol := range cols {
			(*vars)[icol] = (*vars)[name+item+icol]
		}
		out += textproc.Process((*vars)[`for_body`], vars)
	}
	return
}

func ListVal(vars *map[string]string, pars ...string) string {
	if len(pars) != 3 {
		return ``
	}
	return (*vars)[pars[0]+pars[1]+pars[2]]
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
	fmt.Println(`select * from ` + lib.EscapeName(pars[1]) + where)
	value, err := DB.OneRow(`select * from ` + lib.EscapeName(pars[1]) + where).String()
	if err != nil {
		return err.Error()
	}
	for key, val := range value {
		(*vars)[pars[0]+`_`+key] = lib.StripTags(val)
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
	return strings.Replace(lib.StripTags(value), "\n", "\n<br>", -1)
}

func getClass(class string) (string, string) {
	list := strings.Split(class, ` `)

	more := make([]string, 0)
	classes := make([]string, 0)
	for _, ilist := range list {
		if strings.HasPrefix(ilist, `data-`) || strings.IndexByte(ilist, '=') > 0 {
			lr := strings.Split(ilist, `=`)
			if len(lr) == 1 {
				more = append(more, ilist)
			} else if len(lr) == 2 {
				right := strings.Trim(lr[1], `"'`)
				if ok, _ := regexp.MatchString(`(?i)href`, lr[0]); ok {
					if len(right) > 0 && right[0:1] != `#` {
						continue
					}
				}
				more = append(more, fmt.Sprintf(`%s="%s"`, lr[0], right))
			}
		} else if strings.HasPrefix(ilist, `xs-`) || strings.HasPrefix(ilist, `sm-`) ||
			strings.HasPrefix(ilist, `md-`) || strings.HasPrefix(ilist, `lg`) {
			classes = append(classes, `col-`+ilist)
		} else {
			classes = append(classes, ilist)
		}

	}
	return strings.Join(classes, ` `), strings.Join(more, ` `)
}

func getTag(tag string, pars ...string) (out string) {
	if len(pars) == 0 {
		return
	}
	class, more := getClass(pars[0])
	out = fmt.Sprintf(`<%s class="%s" %s>`, tag, class, more)
	for i := 1; i < len(pars); i++ {
		out += pars[i]
	}
	return out + fmt.Sprintf(`</%s>`, tag)
}

func Tag(vars *map[string]string, pars ...string) (out string) {
	var valid bool
	for _, itag := range []string{`h1`, `h2`, `h3`, `h4`, `h5`, `button`, `table`, `thead`, `tbody`, `tr`, `td`} {
		if pars[0] == itag {
			valid = true
			break
		}
	}
	if valid {
		var class, title string
		if len(pars) > 1 {
			title = pars[1]
		}
		if len(pars) > 2 {
			class = lib.Escape(pars[2])
		}
		return fmt.Sprintf(`<%s class="%s">%s</%[1]s>`, pars[0], class, title)
	}
	return ``
}

func Div(vars *map[string]string, pars ...string) (out string) {
	if len((*vars)[`isrow`]) == 0 {
		out = `<div class="row">`
		(*vars)[`isrow`] = `opened`
	}
	out += getTag(`div`, pars...)
	return out
}

func Par(vars *map[string]string, pars ...string) (out string) {
	return getTag(`p`, pars...)
}

func Em(vars *map[string]string, pars ...string) (out string) {
	return getTag(`em`, pars...)
}

func Small(vars *map[string]string, pars ...string) (out string) {
	return getTag(`small`, pars...)
}

func Span(vars *map[string]string, pars ...string) (out string) {
	return getTag(`span`, pars...)
}

func A(vars *map[string]string, pars ...string) (out string) {
	class, more := getClass(pars[0])
	title := ``
	if len(pars) > 1 {
		title = pars[1]
	}
	href := `#`
	if len(pars) > 2 {
		href = pars[2]
	}
	return fmt.Sprintf(`<a class="%s" %s href="%s">%s</a>`, class, more, href, title)
}

func Strong(vars *map[string]string, pars ...string) (out string) {
	return getTag(`strong`, pars...)
}

func Divs(vars *map[string]string, pars ...string) (out string) {
	count := 0

	if len((*vars)[`isrow`]) == 0 {
		out = `<div class="row">`
		(*vars)[`isrow`] = `opened`
	}
	for _, item := range pars {
		class, more := getClass(item)
		out += fmt.Sprintf(`<div class="%s" %s>`, class, more)
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
			//		val = strings.Replace(val, `#!`, `#`, -1)
		}
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

	title := pars[0]
	if len(pars) > 1 {
		title = pars[1]
	}

	if len(pars) < 2 {
		return ``
	}
	if len(pars) >= 3 {
		params = pars[2]
	}
	classParams := ``
	if len(pars) >= 4 {
		//class = pars[3]
		class, more := getClass(pars[3])
		classParams = fmt.Sprintf(`class="%s" %s`, class, more)
	}
	return fmt.Sprintf(`<a onclick="load_template('%s', {%s} )" %s>%s</a>`, pars[0], params, classParams, title)
}

func BtnEdit(vars *map[string]string, pars ...string) string {
	params := ``
	if len(pars) < 2 {
		return ``
	}
	if len(pars) >= 3 {
		params = pars[2]
	}
	return fmt.Sprintf(`<button style="width: 44px;" type="button" class="btn btn-labeled btn-default" onclick="load_template('%s', {%s})"><span class="btn-label"><em class="fa fa-%s"></em></span></button>`,
		lib.Escape(pars[0]), params, lib.Escape(pars[1]))
}

func BlockInfo(vars *map[string]string, pars ...string) string {
	return fmt.Sprintf(`<a href="#" onclick="openBlockDetailPopup('%s')">%[1]s</a>`, pars[0])
}

func Val(vars *map[string]string, pars ...string) string {
	return fmt.Sprintf(`$('#%s').val()`, pars[0])
}

func BtnPage(vars *map[string]string, pars ...string) string {
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
	more := ``
	class := `btn btn-primary`
	if len(pars) >= 4 {
		class, more = getClass(pars[3])
	}
	anchor := `''`
	if len(pars) >= 5 {
		anchor = pars[4]
	}
	return fmt.Sprintf(`<button type="button" class="%s" %s onclick="load_template('%s', {%s}, %s )">%s</button>`,
		class, more, pars[0], params, anchor, pars[1])
}

func BtnContract(vars *map[string]string, pars ...string) string {
	// contract, title, text, params, class, pagetemplate, name, paramssuccess
	params := ``
	onsuccess := ``
	page := ``
	pageparam := ``
	if len(pars) < 3 {
		return ``
	}

	if len(pars) >= 4 {
		params = pars[3]
	}
	if params == `''` {
		params = ``
	}
	class := `"btn btn-primary"`
	if len(pars) >= 5 {
		class = pars[4]
	}
	if len(pars) >= 7 {
		onsuccess = lib.Escape(pars[5])
		page = lib.Escape(pars[6])
		if len(pars) == 8 {
			pageparam = lib.Escape(pars[7])
		}
	}
	(*vars)["wibtncont"] = `1`
	return fmt.Sprintf(`<button type="button" class=%s data-tool="panel-refresh" onclick="btn_contract(this, '%s', {%s}, '%s', '%s', '%s', {%s})">%s</button>`,
		class, pars[0], params, pars[2], onsuccess, page, pageparam, pars[1])
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
	tableClass := ``
	adaptive := ``
	if val, ok := (*pars)[`Order`]; ok {
		order = `order by ` + lib.Escape(val)
	}
	if val, ok := (*pars)[`Class`]; ok {
		tableClass = lib.Escape(val)
	}
	if _, ok := (*pars)[`Adaptive`]; ok {
		adaptive = `data-role="table"`
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

	out := ``
	if strings.TrimSpace(tableClass) == `table-responsive` {
		out += `<div class="table-responsive">`
	}
	out += `<table class="table ` + tableClass + `" ` + adaptive + `><thead>`
	for _, th := range *columns {
		if len(th) < 2 {
			return `incorrect column`
		}
		out += `<th>` + th[0] + `</th>`
		th[1] = strings.TrimSpace(th[1])
		off := strings.Index(th[1], `StateLink`)
		if off >= 0 {
			thname := th[1][off:]
			if strings.IndexByte(thname, ',') > 0 {
				linklist := strings.TrimSpace(thname[strings.IndexByte(thname, '(')+1 : strings.IndexByte(thname, ',')])
				if alist := strings.Split(StateValue(vars, linklist), `,`); len(alist) > 0 {
					for ind, item := range alist {
						(*vars)[fmt.Sprintf(`%s_%d`, linklist, ind+1)] = LangRes(vars, item)
					}
				}
			}
		}
	}
	out += `</thead>`
	for _, item := range list {
		out += `<tr>`
		for key, value := range item {
			if key != `state_id` {
				(*vars)[key] = lib.StripTags(value)
			}
		}
		for _, th := range *columns {
			if len(th) < 2 {
				return `incorrect column`
			}
			//			val := textproc.Process(th[1], vars)
			//			if val == `NULL` {
			val := textproc.Macro(th[1], vars)
			//			}
			out += `<td>` + strings.Replace(val, "\n", "\n<br>", -1) + `</td>`
		}
		out += `</tr>`
	}
	out += `</table>`
	if strings.TrimSpace(tableClass) == `table-responsive` {
		out += `</div>`
	}
	return out
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
	if len(pars) > 1 {
		ind := StrToInt(pars[1])
		if alist := strings.Split(val, `,`); ind > 0 && len(alist) >= ind {
			val = LangRes(vars, alist[ind-1])
		}
	}
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
	return fmt.Sprintf(`<li><a href="#" onclick="load_template('%s', {%s});"><span>%s</span></a></li>`,
		name, params, title)
}

func Navigation(vars *map[string]string, pars ...string) string {
	li := make([]string, 0)
	for _, ipar := range pars {
		li = append(li, ipar)
	}
	return textproc.Macro(fmt.Sprintf(`<ol class="breadcrumb"><span class="pull-right">
	<a href='#' onclick="load_template('sys-editPage', {name: '#page#', global:'#global#'} )">Edit</a></span>%s</ol>`,
		strings.Join(li, `&nbsp;/&nbsp;`)), vars)
}

func MarkDown(vars *map[string]string, pars ...string) string {
	return textproc.Macro(string(blackfriday.MarkdownCommon([]byte(pars[0]))), vars)
}

func Title(vars *map[string]string, pars ...string) string {
	return fmt.Sprintf(`<div class="content-heading">%s</div>`, pars[0])
}

func PageTitle(vars *map[string]string, pars ...string) string {
	var row string
	if len((*vars)[`isrow`]) == 0 {
		row = ` row`
		(*vars)[`isrow`] = `closed`
	}
	return fmt.Sprintf(`<div class="panel panel-default" data-sweet-alert><div class="panel-heading"><div class="panel-title">%s</div></div><div class="panel-body%s">`, pars[0], row)
}

func PageEnd(vars *map[string]string, pars ...string) string {
	if (*vars)[`isrow`] == `closed` {
		(*vars)[`isrow`] = ``
	}
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
	text := LangRes(vars, pars[0])
	return fmt.Sprintf(`<label %s>%s</label>`, class, text)
}

func Legend(vars *map[string]string, pars ...string) (out string) {
	return getTag(`legend`, pars...)
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
	btnName := `Send`
	if btn, ok := (*pars)[`Name`]; ok {
		btnName = btn
	}
	name := (*pars)[`Contract`]
	//	init := (*pars)[`Init`]
	class := `clearfix pull-right`
	//var more, moreBtn string
	if len((*pars)[`Class`]) > 0 {
		class, _ = getClass((*pars)[`Class`])
	}
	classBtn := `btn btn-primary`
	if len((*pars)[`ClassBtn`]) > 0 {
		classBtn, _ = getClass((*pars)[`ClassBtn`])
	}

	onsuccess := (*pars)[`OnSuccess`]
	contract := smart.GetContract(name, uint32(StrToUint64((*vars)[`state_id`])))
	if contract == nil /*|| contract.Block.Info.(*script.ContractInfo).Tx == nil*/ {
		return fmt.Sprintf(`there is not %s contract`, name)
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
	finfo := TxButtonInfo{TxName: name, Class: class, ClassBtn: classBtn, Name: LangRes(vars, btnName),
		Unique: template.JS((*vars)[`tx_unique`]), OnSuccess: template.JS(onsuccess),
		Fields: make([]TxInfo, 0), AutoClose: (*pars)[`AutoClose`] != `0`,
		Silent: (*pars)[`Silent`] == `1`, Data: FormCommon{CountSignArr: []byte{1}}}

	idnames := strings.Split((*pars)[`Inputs`], `,`)
	names := make(map[string]string)
	values := make(map[string]string)
	for _, idn := range idnames {
		if lr := strings.SplitN(idn, `#=`, 2); len(lr) == 2 {
			values[strings.TrimSpace(lr[0])] = strings.TrimSpace(lr[1])
		} else {
			if lr = strings.SplitN(idn, `=`, 2); len(lr) == 2 {
				names[strings.TrimSpace(lr[0])] = strings.TrimSpace(lr[1])
			}
		}
	}
	if contract.Block.Info.(*script.ContractInfo).Tx != nil {
	txlist:
		for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
			value := ``
			idname := fitem.Name
			if idn, ok := values[idname]; ok {
				value = (*vars)[idn]
				idname = idn + (*vars)[`tx_unique`]
			} else if idn, ok = names[idname]; ok {
				idname = idn
			}
			for _, tag := range []string{`date`, `polymap`, `map`, `image`, `text`, `address`} {
				if strings.Index(fitem.Tags, tag) >= 0 {
					finfo.Fields = append(finfo.Fields, TxInfo{Name: fitem.Name, Id: idname, Value: value, HtmlType: tag})
					continue txlist
				}
			}
			if fitem.Type.String() == `decimal.Decimal` {
				var count int
				if ret := regexp.MustCompile(`(?is)digit:(\d+)`).FindStringSubmatch(fitem.Tags); len(ret) == 2 {
					count = StrToInt(ret[1])
				} else {
					count = StrToInt(StateValue(vars, `money_digit`))
				}
				finfo.Fields = append(finfo.Fields, TxInfo{Name: fitem.Name, Value: value, HtmlType: "money",
					Id: idname, Param: IntToStr(count)})
			} else {
				finfo.Fields = append(finfo.Fields, TxInfo{Name: fitem.Name, Value: value, Id: idname, HtmlType: "textinput"})
			}
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
	finfo := FormInfo{TxName: name, Unique: template.JS((*vars)[`tx_unique`]), OnSuccess: template.JS(onsuccess),
		Fields: make([]FieldInfo, 0), AutoClose: (*pars)[`AutoClose`] != `0`,
		Silent: (*pars)[`Silent`] == `1`, Data: FormCommon{CountSignArr: []byte{1}}}

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
txlist:
	for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
		var value string
		if val, ok := (*vars)[fitem.Name]; ok {
			value = val
		}
		if strings.Index(fitem.Tags, `hidden`) >= 0 || strings.Index(fitem.Tags, `signature`) >= 0 {
			finfo.Fields = append(finfo.Fields, FieldInfo{Name: fitem.Name, HtmlType: `hidden`,
				TxType: fitem.Type.String(), Title: ``, Value: value})
			continue
		}
		langres := gettag('#', fitem.Name, fitem.Tags)
		linklist := gettag('@', ``, fitem.Tags)
		title := LangRes(vars, langres)
		for _, tag := range []string{`date`, `polymap`, `map`, `image`, `text`, `address`} {
			if strings.Index(fitem.Tags, tag) >= 0 {
				if tag == `date` {
					(*vars)[`widate`] = `1`
				}
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
						sellist.List[int(StrToInt64(item[id]))] = lib.StripTags(item[name])
					}
				}
			} else if alist := strings.Split(StateValue(vars, linklist), `,`); len(alist) > 0 {
				for ind, item := range alist {
					sellist.List[ind+1] = LangRes(vars, item)
				}
			}
			finfo.Fields = append(finfo.Fields, FieldInfo{Name: fitem.Name, HtmlType: "select",
				TxType: fitem.Type.String(), Title: title, Value: sellist})
		} else if fitem.Type.String() == `decimal.Decimal` {
			var count int
			if ret := regexp.MustCompile(`(?is)digit:(\d+)`).FindStringSubmatch(fitem.Tags); len(ret) == 2 {
				count = StrToInt(ret[1])
			} else {
				count = StrToInt(StateValue(vars, `money_digit`))
			}
			value = Money(vars, value, IntToStr(count))
			finfo.Fields = append(finfo.Fields, FieldInfo{Name: fitem.Name, HtmlType: "money",
				TxType: fitem.Type.String(), Title: title, Value: value,
				Param: IntToStr(count) /*`9{1,20}` + postfix*/})
		} else if fitem.Type.String() == `string` || fitem.Type.String() == `int64` || fitem.Type.String() == `float64` {
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
	count := 0
	size := 18
	if len(pars) > 0 {
		count = int(StrToInt64(pars[0]))
	}
	if len(pars) > 1 {
		size = int(StrToInt64(pars[1]))
	}
	pct := 100
	if len(pars) > 2 {
		pct = int(StrToInt64(pars[2]))
	}
	speed := 1
	if len(pars) > 3 {
		speed = int(StrToInt64(pars[3]))
	}
	color := `23b7e5`
	if len(pars) > 4 {
		color = pars[4]
	}
	fontColor := `656565`
	if len(pars) > 5 {
		fontColor = pars[5]
	}
	width := 250
	if len(pars) > 6 {
		width = int(StrToInt64(pars[6]))
	}
	thickness := 10
	if len(pars) > 7 {
		thickness = int(StrToInt64(pars[7]))
	}
	prefix := ``
	if len(pars) > 8 {
		prefix = pars[8]
	}
	suffix := ``
	if len(pars) > 9 {
		suffix = pars[9]
	}
	return fmt.Sprintf(`
		<div
                    data-count
                    data-count-font="%dpx"
                    data-count-number="%d"
                    data-count-percentage="%d"
                    data-count-speed="%d"
                    data-count-color="#%s"
                    data-count-font-color="#%s"
                    data-count-width="%d"
                    data-count-thickness="%d"
                    data-count-prefix="%s"
                    data-count-suffix="%s"
                    data-count-outline="rgba(200,200,200,0.4)"
                    data-count-fill="#ffffff"
                    data-count-pie=""
                    data-count-separator=" "
                    data-count-decimal=" "
                    data-count-decimals=" "
                ></div>`, size, count, pct, speed, color, fontColor, width, thickness, prefix, suffix)
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

func Source(vars *map[string]string, pars ...string) string {
	var value string
	if len(pars) > 1 {
		value = pars[1]
	}
	(*vars)["wisource"] = pars[0]
	return fmt.Sprintf(`<pre class="textEditor"><code></code><section id="textEditor">%s</section>
					</pre>
				   <textarea id="%s" class="form-control hidden"></textarea>`, value, pars[0])
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
		format = LangRes(vars, `dateformat`)
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
		format = LangRes(vars, `timeformat`)
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
					list = append(list, SelInfo{Id: StrToInt64(item[id]), Name: lib.StripTags(item[name])})
				}
			}
		} else if alist := strings.Split(StateValue(vars, pars[1]), `,`); len(alist) > 0 {
			for ind, item := range alist {
				list = append(list, SelInfo{Id: int64(ind + 1), Name: LangRes(vars, item)})
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
	class := ``
	if len(pars) > 1 {
		class = pars[1]
	}
	(*vars)[`wimap`] = `1`
	return fmt.Sprintf(`<div class="wimap %s">%s</div>`, class, pars[0])
}

func MapPoint(vars *map[string]string, pars ...string) string {
	(*vars)[`wimappoint`] = `1`
	return fmt.Sprintf(`<div class="wimappoint">%s</div>`, pars[0])
}

func MenuGroup(vars *map[string]string, pars ...string) string {
	var (
		idname, icon string
	)
	/*	id = (*vars)[`menuid`]
		if len(id) > 0 {
			id = IntToStr(StrToInt(id) + 1)
		} else {
			id = `0`
		}
		(*vars)[`menuid`] = id*/
	if len(pars) > 1 {
		idname = lib.Escape(pars[1])
	}
	if len(pars) > 2 {
		icon = fmt.Sprintf(`<em class="%s"></em>`, lib.Escape(pars[2]))
	}
	return fmt.Sprintf(`<li id="li%s"><span>%s
     <span>%s</span></span>
	 <ul id="ul%[1]s">`,
		idname, icon, LangRes(vars, lib.Escape(pars[0])))
}

func MenuItem(vars *map[string]string, pars ...string) string {
	var (
		/*idname,*/ action, page, params, icon string
	)
	/*	if len(pars) > 1 {
		idname = lib.Escape(pars[1])
	}*/
	off := 0
	if len(pars) > 1 {
		action = lib.Escape(pars[1])
	}
	if !strings.HasPrefix(action, `load_`) {
		action = `load_template`
		off = 1
	}
	if len(pars) > 2-off {
		page = lib.Escape(pars[2-off])
	}
	if len(pars) > 3-off {
		params = lib.Escape(pars[3-off])
	}
	if len(pars) > 4-off {
		icon = fmt.Sprintf(`<em class="%s"></em>`, lib.Escape(pars[4-off]))
	}
	return fmt.Sprintf(`<li id="li%s">
		<a href="#" title="%s" onClick="%s('%s',{%s});">
		%s<span>%[2]s</span></a></li>`,
		page, LangRes(vars, lib.Escape(pars[0])), action, page, params, icon)
}

func MenuPage(vars *map[string]string, pars ...string) string {
	return fmt.Sprintf(`<!--%s-->`, lib.Escape(pars[0]))
}

func MenuBack(vars *map[string]string, pars ...string) string {
	var link string
	if len(pars) > 1 {
		link = fmt.Sprintf(`load_template('%s')`, lib.Escape(pars[1]))
	}
	return fmt.Sprintf(`<!--%s=%s-->`, lib.Escape(pars[0]), link)
}

func MenuEnd(vars *map[string]string, pars ...string) string {
	return `</ul></li>`
}

func ChartBar(vars *map[string]string, pars *map[string]string) string {
	id := fmt.Sprintf(`bar%d`, RandInt(0, 0xfffffff))
	data := make([]string, 0)
	labels := make([]string, 0)
	//	if len((*pars)[`Data`]) > 0 {
	//	} else {
	colors := strings.Split((*pars)[`Colors`], `,`)
	value := (*pars)[`FieldValue`]
	label := (*pars)[`FieldLabel`]
	if len(colors) == 0 {
		colors = []string{`23b7e5`}
	}
	if len(value) == 0 || len(label) == 0 {
		return `empty FieldValue or FieldLabel`
	}
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
	list, err := DB.GetAll(fmt.Sprintf(`select %s,%s from %s %s %s%s`, lib.EscapeName(value), lib.EscapeName(label),
		lib.EscapeName((*pars)[`Table`]), where, order, limit), -1)
	if err != nil {
		return err.Error()
	}
	for _, item := range list {
		data = append(data, lib.StripTags(item[value]))
		labels = append(labels, `'`+lib.StripTags(item[label])+`'`)
	}
	//	}
	return fmt.Sprintf(`<div><canvas id="%s"></canvas>
		</div><script language="JavaScript" type="text/javascript">
		(function (){
    var barData = {
        labels : [%s],
        datasets : [
          {
            fillColor : '#%[3]s',
            strokeColor : '#%[3]s',
            highlightFill: '#%[3]s',
            highlightStroke: '#%[3]s',
            data : [%s]
          }
        ]
    };
    
    var barOptions = {
      scaleBeginAtZero : true,
      scaleShowGridLines : true,
      scaleGridLineColor : 'rgba(0,0,0,.05)',
      scaleGridLineWidth : 1,
      barShowStroke : true,
      barStrokeWidth : 2,
      barValueSpacing : 5,
      barDatasetSpacing : 1,
      responsive: true
    };

    var barctx = document.getElementById("%s").getContext("2d");
    var barChart = new Chart(barctx).Bar(barData, barOptions);
	})();
</script>`, id, strings.Join(labels, ","), colors[0], strings.Join(data, ","), id)
}

func ChartPie(vars *map[string]string, pars *map[string]string) string {
	id := fmt.Sprintf(`pie%d`, RandInt(0, 0xfffffff))
	out := make([]string, 0)

	if len((*pars)[`Data`]) > 0 {
		data := textproc.Split((*pars)[`Data`])
		for _, item := range *data {
			if len(item) == 3 {
				out = append(out, fmt.Sprintf(`{
				value: %s,
				color: '#%s',
				highlight: '#%s',
				label: '%s'
				}`, item[0], item[1], item[1], item[2]))
			}
		}
	} else {
		colors := strings.Split((*pars)[`Colors`], `,`)
		value := (*pars)[`FieldValue`]
		label := (*pars)[`FieldLabel`]
		if len(colors) == 0 {
			return `empty color parameter in ChartPie`
		}
		if len(value) == 0 || len(label) == 0 {
			return `empty FieldValue or FieldLabel`
		}
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
		} else {
			limit = fmt.Sprintf(` limit %d`, len(colors))
		}
		list, err := DB.GetAll(fmt.Sprintf(`select %s,%s from %s %s %s%s`, lib.EscapeName(value), lib.EscapeName(label),
			lib.EscapeName((*pars)[`Table`]), where, order, limit), -1)
		if err != nil {
			return err.Error()
		}
		for ind, item := range list {
			color := colors[ind%len(colors)]
			out = append(out, fmt.Sprintf(`{
				value: %s,
				color: '#%s',
				highlight: '#%s',
				label: '%s'
			}`, lib.StripTags(item[value]), color, color, lib.StripTags(item[label])))
		}
	}
	return fmt.Sprintf(`<div><canvas id="%s"></canvas>
		</div><script language="JavaScript" type="text/javascript">
		(function (){
    var pieData =[
          %s
        ];

    var pieOptions = {
      segmentShowStroke : true,
      segmentStrokeColor : '#fff',
      segmentStrokeWidth : 2,
      percentageInnerCutout : 0, 
      animationSteps : 100,
      animationEasing : 'easeOutBounce',
      animateRotate : true,
      animateScale : false,
      responsive: true
    };

    var piectx = document.getElementById("%s").getContext("2d");
    var pieChart = new Chart(piectx).Pie(pieData, pieOptions);
	})();
</script>`, id, strings.Join(out, ",\r\n"), id)
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
		"replaceBr": func(text string) template.HTML {
			text = strings.Replace(text, `\n`, "<br>", -1)
			text = strings.Replace(text, `\t`, "&nbsp;&nbsp;&nbsp;&nbsp;", -1)
			return template.HTML(text)
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
