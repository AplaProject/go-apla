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

package template

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/language"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/static"
	"github.com/EGaaS/go-egaas-mvp/packages/textproc"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/sql"
	"github.com/op/go-logging"
	"github.com/russross/blackfriday"
	"github.com/shopspring/decimal"
)

var log = logging.MustGetLogger("daemons")

// FieldInfo contains the information of contract data field
type FieldInfo struct {
	Name     string      `json:"name"`
	HTMLType string      `json:"htmlType"`
	TxType   string      `json:"txType"`
	Title    string      `json:"title"`
	Value    interface{} `json:"value"`
	Param    string      `json:"param"`
}

/*type FormCommon struct {
	//Lang   map[string]string
		Address      string
		WalletId     int64
		CitizenId    int64
		StateId      int64
		StateName    string
}*/

// FormInfo contains parameters of TxForm function
type FormInfo struct {
	TxName    string
	Unique    template.JS
	OnSuccess template.JS
	Fields    []FieldInfo
	AutoClose bool
	Silent    bool
	//Data      FormCommon
}

// TxInfo contains the information of contract data field in TxButton function
type TxInfo struct {
	Name     string `json:"name"`
	ID       string `json:"id"`
	Value    string `json:"value"`
	HTMLType string `json:"htmlType"`
	Param    string `json:"param"`
}

// TxButtonInfo contains parameters of TxButton function
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
}

// TxBtnCont contains parameters of TxBtnCont function
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
}

// CommonPage contains the common information for each template page
type CommonPage struct {
	Address   string
	WalletId  int64
	CitizenId int64
	StateId   int64
	StateName string
}

// PageTpl is the main structure for the template page
type PageTpl struct {
	Page     string
	Template string
	Unique   string
	Data     interface{} //*CommonPage
}

// SelList is a structure for selectable data
type SelList struct {
	Cur  int64          `json:"cur"`
	List map[int]string `json:"list"`
}

// SelInfo is a structure for an item of selectable data
type SelInfo struct {
	ID   int64
	Name string
}

func init() {
	smart.Extend(&script.ExtendData{Objects: map[string]interface{}{
		"Balance":    Balance,
		"StateParam": StateParam,
		/*		"DBInsert":   DBInsert,
		 */
	}, AutoPars: map[string]string{
	//		`*parser.Parser`: `parser`,
	}})

	textproc.AddMaps(&map[string]textproc.MapFunc{`Table`: Table, `TxForm`: TxForm, `TxButton`: TXButton,
		`ChartPie`: ChartPie, `ChartBar`: ChartBar})
	textproc.AddFuncs(&map[string]textproc.TextFunc{`Address`: IDToAddress, `BtnEdit`: BtnEdit,
		`InputMap`: InputMap, `InputMapPoly`: InputMapPoly,
		`Image`: Image, `ImageInput`: ImageInput, `Div`: Div, `P`: Par, `Em`: Em, `Small`: Small, `A`: A, `Span`: Span, `Strong`: Strong, `Divs`: Divs, `DivsEnd`: DivsEnd,
		`LiTemplate`: LiTemplate, `LinkPage`: LinkPage, `BtnPage`: BtnPage, `UList`: UList, `UListEnd`: UListEnd, `Li`: Li,
		`CmpTime`: CmpTime, `Title`: Title, `MarkDown`: MarkDown, `Navigation`: Navigation, `PageTitle`: PageTitle,
		`PageEnd`: PageEnd, `StateVal`: StateVal, `Json`: JSONScript, `And`: And, `Or`: Or, `LiBegin`: LiBegin, `LiEnd`: LiEnd,
		`TxId`: TxID, `SetVar`: SetVar, `GetList`: GetList, `GetRow`: GetRowVars, `GetOne`: GetOne, `TextHidden`: TextHidden,
		`ValueById`: ValueByID, `FullScreen`: FullScreen, `Ring`: Ring, `WiBalance`: WiBalance, `GetVar`: GetVar,
		`WiAccount`: WiAccount, `WiCitizen`: WiCitizen, `Map`: Map, `MapPoint`: MapPoint, `StateLink`: StateLink,
		`If`: If, `IfEnd`: IfEnd, `Else`: Else, `ElseIf`: ElseIf, `Trim`: Trim, `Date`: Date, `DateTime`: DateTime, `Now`: Now, `Input`: Input,
		`Textarea`: Textarea, `InputMoney`: InputMoney, `InputAddress`: InputAddress, `ForList`: ForList, `ForListEnd`: ForListEnd,
		`BlockInfo`: BlockInfo, `Back`: Back, `ListVal`: ListVal, `Tag`: Tag, `BtnContract`: BtnContract,
		`Form`: Form, `FormEnd`: FormEnd, `Label`: Label, `Legend`: Legend, `Select`: Select, `Param`: Param, `Mult`: Mult,
		`Money`: Money, `Source`: Source, `Val`: Val, `Lang`: LangRes, `LangJS`: LangJS, `InputDate`: InputDate,
		`MenuGroup`: MenuGroup, `MenuEnd`: MenuEnd, `MenuItem`: MenuItem, `MenuPage`: MenuPage, `MenuBack`: MenuBack,
		`WhiteMobileBg`: WhiteMobileBg, `Bin2Hex`: Bin2Hex, `MessageBoard`: MessageBoard, `AutoUpdate`: AutoUpdate,
		`AutoUpdateEnd`: AutoUpdateEnd, `Include`: Include,
	})
}

// LoadContracts reads and compiles contracts from smart_contracts tables
func LoadContracts() (err error) {
	var states []map[string]string
	prefix := []string{`global`}
	states, err = sql.DB.GetAll(`select id from system_states order by id`, -1)
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

// LoadContract reads and compiles contract of new state
func LoadContract(prefix string) (err error) {
	var contracts []map[string]string
	contracts, err = sql.DB.GetAll(`select * from "`+prefix+`_smart_contracts" order by id`, -1)
	if err != nil {
		return err
	}
	for _, item := range contracts {
		if err = smart.Compile(item[`value`], prefix, item[`active`] == `1`, converter.StrToInt64(item[`id`])); err != nil {
			log.Error("Load Contract", item[`name`], err)
			fmt.Println("Error Load Contract", item[`name`], err)
			//return
		} else {
			fmt.Println("OK Load Contract", item[`name`], item[`id`], item[`active`] == `1`)
		}
	}
	return
}

// Balance returns the balance of the wallet
func Balance(walletID int64) (decimal.Decimal, error) {
	balance, err := sql.DB.Single("SELECT amount FROM dlt_wallets WHERE wallet_id = ?", walletID).String()
	if err != nil {
		return decimal.New(0, 0), err
	}
	return decimal.NewFromString(balance)
}

// EGSRate returns egs_rate of the state
func EGSRate(idstate int64) (float64, error) {
	return sql.DB.Single(`SELECT value FROM "`+converter.Int64ToStr(idstate)+`_state_parameters" WHERE name = ?`, `egs_rate`).Float64()
}

// StateParam returns the value of state parameters
func StateParam(idstate int64, name string) (string, error) {
	return sql.DB.Single(`SELECT value FROM "`+converter.Int64ToStr(idstate)+`_state_parameters" WHERE name = ?`, name).String()
}

// Param returns the value of the specified varaible
func Param(vars *map[string]string, pars ...string) string {
	if val, ok := (*vars)[pars[0]]; ok {
		return val
	}
	return ``
}

// LangRes returns the corresponding language resource of the specified parameter
func LangRes(vars *map[string]string, pars ...string) string {
	ret, _ := language.LangText(pars[0], int(converter.StrToInt64((*vars)[`state_id`])), (*vars)[`accept_lang`])
	return ret
}

// LangJS returns span tag for the language resource
func LangJS(vars *map[string]string, pars ...string) string {
	return fmt.Sprintf(`<span class="lang" lang-id="%s"></span>`, pars[0])
}

func ifValue(val string) bool {
	var sep string
	if strings.Index(val, `;base64`) < 0 {
		for _, item := range []string{`==`, `!=`, `<=`, `>=`, `<`, `>`} {
			if strings.Index(val, item) >= 0 {
				sep = item
				break
			}
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

// Money returns the formated value of the specified money amount
func Money(vars *map[string]string, pars ...string) string {
	var cents int
	if len(pars) > 1 {
		cents = converter.StrToInt(pars[1])
	} else {
		cents = converter.StrToInt(StateVal(vars, `money_digit`))
	}
	ret := pars[0]
	if ret == `NULL` {
		ret = `0`
	}
	if cents > 0 && strings.IndexByte(ret, '.') < 0 {
		if len(ret) < cents+1 {
			ret = strings.Repeat(`0`, cents+1-len(ret)) + ret
		}
		ret = ret[:len(ret)-cents] + `.` + ret[len(ret)-cents:]
	}
	return ret
}

// And is a logical AND function
func And(vars *map[string]string, pars ...string) string {
	for _, item := range pars {
		if !ifValue(item) {
			return `0`
		}
	}
	return `1`
}

// Or is a logical OR function
func Or(vars *map[string]string, pars ...string) string {
	for _, item := range pars {
		if ifValue(item) {
			return `1`
		}
	}
	return `0`
}

// CmpTime compares two time. It returns 0 if they equal, -1 - left < right, 1 - left > right
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

// If function emulates conditional operator for text processing
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

// Else function emulates else in conditional operator for text processing
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

// ElseIf function emulates 'else if' for text processing
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

// IfEnd must be used at the end of If operator for text processing
func IfEnd(vars *map[string]string, pars ...string) string {
	ilen := len((*vars)[`ifs`])
	if ilen > 0 {
		(*vars)[`ifs`] = (*vars)[`ifs`][:ilen-1]
	}
	return ``
}

// Now returns the current time of postgresql
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
	ret, err := sql.DB.Single(query).String()
	if err != nil {
		return err.Error()
	}
	if cut > 0 {
		ret = strings.Replace(ret[:cut], `T`, ` `, -1)
	}
	return ret
}

// Textarea returns textarea HTML tag
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

// Input returns input HTML tag
func Input(vars *map[string]string, pars ...string) string {
	var (
		class, value, more, placeholder string
	)
	itype := `text`
	if len(pars) > 1 {
		class, more = getClass(pars[1])
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
	return fmt.Sprintf(`<input type="%s" id="%s" placeholder="%s" class="%s" value="%s" %s>`,
		itype, pars[0], placeholder, class, value, more)
}

// InputDate returns input HTML tag with datepicker
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

// InputMoney returns input HTML tag with a special money mask
func InputMoney(vars *map[string]string, pars ...string) string {
	var (
		class, value string
		digit        int
	)
	if len(pars) > 1 {
		class = pars[1]
	}
	if len(pars) > 3 {
		digit = converter.StrToInt(pars[3])
	} else {
		digit = converter.StrToInt(StateVal(vars, `money_digit`))
	}
	if len(pars) > 2 {
		value = Money(vars, pars[2], converter.IntToStr(digit))
	}
	(*vars)["wimoney"] = `1`
	return fmt.Sprintf(`<input id="%s" type="text" value="%s"
				data-inputmask="'alias': 'numeric', 'rightAlign': false, 'groupSeparator': ' ', 'autoGroup': true, 'digits': %d, 'digitsOptional': false, 'prefix': '', 'placeholder': '0'"
	class="inputmask %s">`, pars[0], value, digit, class)
}

// InputAddress returns input HTML tag for entering wallet address
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

// Trim trims spaces at the beginning and at the end of the text
func Trim(vars *map[string]string, pars ...string) string {
	if len(pars) == 0 {
		return ``
	}
	return strings.TrimSpace(pars[0])
}

// Back returns back button
func Back(vars *map[string]string, pars ...string) string {
	if len(pars[0]) == 0 || len(pars) < 2 || len(pars[1]) == 0 {
		return ``
	}
	var params string
	if len(pars) == 3 {
		params = converter.Escape(pars[2])
	}
	return fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
	hist_push(['load_%s', '%s', {%s}]);
</script>`, converter.Escape(pars[0]), converter.Escape(pars[1]), params)
}

// JSONScript returns json object
func JSONScript(vars *map[string]string, pars ...string) string {
	if len(pars) == 0 {
		return ``
	}
	return fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
	var jdata = { 
%s 
}
</script>`, pars[0])
}

// FullScreen inserts java script for switching the workarea to the full browser window
func FullScreen(vars *map[string]string, pars ...string) string {
	wide := `add`
	if len(pars) > 0 && pars[0] == `0` {
		wide = `remove`
	}
	return fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
	$("body").%sClass('wide');
</script>`, wide)
}

// WhiteMobileBg switches flatPageMobile class
func WhiteMobileBg(vars *map[string]string, pars ...string) string {
	wide := `add`
	if len(pars) > 0 && pars[0] == `0` {
		wide = `remove`
	}
	return fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
	$("body").%sClass('flatPageMobile');
</script>`, wide)
}

// Bin2Hex converts interface to hex string
func Bin2Hex(vars *map[string]string, pars ...string) string {
	if len(pars) == 0 {
		return ``
	}
	return string(converter.BinToHex(pars[0]))
}

// WhiteBg switches flatPageMobile class
func WhiteBg(vars *map[string]string, pars ...string) string {
	wide := `add`
	if len(pars) > 0 && pars[0] == `0` {
		wide = `remove`
	}
	return fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
	$("body").%sClass('flatPage');
</script>`, wide)
}

// MessageBoard returns HTML source for displaying messages
func MessageBoard(vars *map[string]string, pars ...string) string {
	messages, err := sql.DB.GetAll(`select * from "global_messages" order by id`, 100)
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

// GetList assigns the result of sql request to the variables
func GetList(vars *map[string]string, pars ...string) string {
	// name, table, fields, where, order, limit
	if len(pars) < 3 {
		return ``
	}
	where := ``
	order := ``
	limit := -1
	fields := converter.Escape(pars[2])
	keys := strings.Split(fields, `,`)
	if len(pars) >= 4 {
		where = ` where ` + converter.Escape(pars[3])
	}
	if len(pars) >= 5 {
		order = ` order by ` + converter.EscapeName(pars[4])
	}
	if len(pars) >= 6 {
		limit = converter.StrToInt(pars[5])
	}

	value, err := sql.DB.GetAll(`select `+fields+` from `+converter.EscapeName(pars[1])+where+order, limit)
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
				ival = converter.StripTags(ival)
			}
			if ival == `NULL` {
				ival = ``
			}
			(*vars)[pars[0]+ikey+key] = ival
		}
		//		out[item[keys[0]]] = item
	}
	(*vars)[pars[0]+`_list`] = strings.Join(list, `|`)
	(*vars)[pars[0]+`_columns`] = strings.Join(cols, `|`)
	return ``
}

// AutoUpdate reloads inner commands each pars[0] seconds
func AutoUpdate(vars *map[string]string, pars ...string) string {
	time := converter.StrToInt(pars[0])
	if time == 0 {
		time = 10
	}
	(*vars)[`auto_time`] = converter.IntToStr(time)
	(*vars)[`auto_loop`] = `1`
	if len((*vars)[`auto_id`]) > 0 {
		(*vars)[`auto_id`] = converter.IntToStr(converter.StrToInt((*vars)[`auto_id`]) + 1)
	} else {
		(*vars)[`auto_id`] = `1`
	}
	return fmt.Sprintf(`<div id="auto%s">`, (*vars)[`auto_id`])
}

// AutoUpdateEnd must be used with AutoUpdate for text processing
func AutoUpdateEnd(vars *map[string]string, pars ...string) (out string) {
	out = fmt.Sprintf(`</div><div id="auto%sbody" style="display:none;">%s</div>
<script language="JavaScript" type="text/javascript">
setTimeout( function(){ autoUpdate(%[1]s, %[3]s); }, %[3]s000 );
</script>`, (*vars)[`auto_id`], (*vars)[`auto_body`], (*vars)[`auto_time`])
	return
}

// Include returns the another template page
func Include(vars *map[string]string, pars ...string) string {
	params := make(map[string]string)
	for i, val := range pars {
		if i > 0 {
			lr := strings.SplitN(val, `=`, 2)
			if len(lr) == 2 {
				params[lr[0]] = lr[1]
			}
		}
	}
	//	page := (*vars)[`page`]
	out, err := CreateHTMLFromTemplate(pars[0], converter.StrToInt64((*vars)[`citizen`]), converter.StrToInt64((*vars)[`state_id`]),
		&params)
	if err != nil {
		out = err.Error()
	}
	//	(*vars)[`page`] = page
	return out
}

// ForList emulates for operator in text processing
func ForList(vars *map[string]string, pars ...string) string {
	(*vars)[`for_name`] = pars[0]
	(*vars)[`for_loop`] = `1`
	return ``
}

// ForListEnd must be used with ForList for text processing
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

// ListVal returns the value of the list as the value of the variable
func ListVal(vars *map[string]string, pars ...string) string {
	if len(pars) != 3 {
		return ``
	}
	return (*vars)[pars[0]+pars[1]+pars[2]]
}

// GetRowVars assignes the value of row result to the variables
func GetRowVars(vars *map[string]string, pars ...string) string {
	if len(pars) != 4 && len(pars) != 3 {
		return ``
	}
	where := ``
	if len(pars) == 4 {
		where = ` where ` + converter.EscapeName(pars[2]) + `='` + converter.Escape(pars[3]) + `'`
	} else if len(pars) == 3 {
		where = ` where ` + converter.Escape(pars[2])
	}
	fmt.Println(`select * from ` + converter.EscapeName(pars[1]) + where)
	value, err := sql.DB.OneRow(`select * from ` + converter.EscapeName(pars[1]) + where).String()
	if err != nil {
		return err.Error()
	}
	for key, val := range value {
		if val == `NULL` {
			val = ``
		}
		(*vars)[pars[0]+`_`+key] = converter.StripTags(val)
	}
	return ``
}

// GetOne returns the single value of sql query.
func GetOne(vars *map[string]string, pars ...string) string {
	if len(pars) < 2 {
		return ``
	}
	where := ``
	if len(pars) == 4 {
		where = ` where ` + converter.EscapeName(pars[2]) + `='` + converter.Escape(pars[3]) + `'`
	} else if len(pars) == 3 {
		where = ` where ` + converter.Escape(pars[2])
	}
	value, err := sql.DB.Single(`select ` + converter.Escape(pars[0]) + ` from ` + converter.EscapeName(pars[1]) + where).String()
	if err != nil {
		return err.Error()
	}
	if value == `NULL` {
		value = ``
	}
	return strings.Replace(converter.StripTags(value), "\n", "\n<br>", -1)
}

func getClass(class string) (string, string) {
	//list := strings.Split(class, ` `)
	list := make([]string, 0)
	buf := make([]rune, 0)
	var quote bool
	for _, ch := range class {
		if ch == ' ' && !quote && len(buf) > 0 {
			list = append(list, string(buf))
			buf = buf[:0]
		} else {
			if ch == '"' {
				quote = !quote
			}
			buf = append(buf, ch)
		}
	}
	if len(buf) > 0 {
		list = append(list, string(buf))
	}
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
		} else if ilist != `''` {
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

// Tag returns the specified HTML tag
func Tag(vars *map[string]string, pars ...string) (out string) {
	var valid bool
	for _, itag := range []string{`h1`, `h2`, `h3`, `h4`, `h5`, `div`, `button`, `table`, `thead`, `tbody`, `tr`, `td`} {
		if pars[0] == itag {
			valid = true
			break
		}
	}
	if valid {
		var class, title, more string
		if len(pars) > 1 {
			title = pars[1]
		}
		if len(pars) > 2 {
			class, more = getClass(converter.Escape(pars[2]))
		}
		return fmt.Sprintf(`<%s class="%s" %s>%s</%[1]s>`, pars[0], class, more, title)
	}
	return ``
}

// Div returns div HTML tag
func Div(vars *map[string]string, pars ...string) (out string) {
	if len((*vars)[`isrow`]) == 0 && (*vars)[`auto_loop`] != `1` {
		out = `<div class="row">`
		(*vars)[`isrow`] = `opened`
	}
	out += getTag(`div`, pars...)
	return out
}

// Par returns paragraph HTML tag
func Par(vars *map[string]string, pars ...string) (out string) {
	return getTag(`p`, pars...)
}

// Em returns em HTML tag
func Em(vars *map[string]string, pars ...string) (out string) {
	return getTag(`em`, pars...)
}

// Li returns li HTML tag
func Li(vars *map[string]string, pars ...string) (out string) {
	class := ``
	more := ``
	if val, ok := (*vars)[`uls`]; ok {
		class = (*vars)[`liclass`+val]
		more = (*vars)[`limore`+val]
	}
	if len(pars) > 1 {
		class, more = getClass(pars[1])
	}
	return fmt.Sprintf(`<li class="%s" %s>%s</li>`, class, more, pars[0])
}

// Small returns small HTML tag
func Small(vars *map[string]string, pars ...string) (out string) {
	return getTag(`small`, pars...)
}

// Span returns span HTML tag
func Span(vars *map[string]string, pars ...string) (out string) {
	return getTag(`span`, pars...)
}

// A returns href HTML tag
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

// Strong returns strong HTML tag
func Strong(vars *map[string]string, pars ...string) (out string) {
	return getTag(`strong`, pars...)
}

// Divs returns nested div HTML tags
func Divs(vars *map[string]string, pars ...string) (out string) {
	count := 0

	if len((*vars)[`isrow`]) == 0 && (*vars)[`auto_loop`] != `1` {
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

// DivsEnd closes divs which has been created with Divs function
func DivsEnd(vars *map[string]string, pars ...string) (out string) {
	if val, ok := (*vars)[`divs`]; ok && len(val) > 0 {
		divs := strings.Split(val, `,`)
		out = strings.Repeat(`</div>`, converter.StrToInt(divs[len(divs)-1]))
		(*vars)[`divs`] = strings.Join(divs[:len(divs)-1], `,`)
	}
	return
}

func tagOut(tag, class, more string) string {
	return fmt.Sprintf(`<%s class="%s" %s>`, tag, class, more)
}

// UList creates ol or ul HTML tag
func UList(vars *map[string]string, pars ...string) string {
	var liclass, limore string

	tag := `ul`
	if len(pars) > 1 && pars[1] == `ol` {
		tag = `ol`
	}
	if val, ok := (*vars)[`uls`]; ok {
		(*vars)[`uls`] = val + tag[:1]
	} else {
		(*vars)[`uls`] = tag[:1]
	}
	class, more := getClass(pars[0])
	if len(pars) > 2 {
		liclass, limore = getClass(pars[2])
	}
	(*vars)[`liclass`+(*vars)[`uls`]] = liclass
	(*vars)[`limore`+(*vars)[`uls`]] = limore
	return tagOut(tag, class, more)
}

// UListEnd closes ol or ul HTMl tag
func UListEnd(vars *map[string]string, pars ...string) string {
	tag := `ul`
	ulen := len((*vars)[`uls`])
	if ulen == 0 {
		return ``
	}
	if (*vars)[`uls`][ulen-1] == 'o' {
		tag = `ol`
	}
	(*vars)[`uls`] = (*vars)[`uls`][:ulen-1]
	return fmt.Sprintf(`</%s>`, tag)
}

// LiBegin opens li HTML tag
func LiBegin(vars *map[string]string, pars ...string) string {
	class, more := getClass(pars[0])
	if val, ok := (*vars)[`uls`]; ok {
		class = (*vars)[`liclass`+val]
		more = (*vars)[`limore`+val]
	}
	return tagOut(`li`, class, more)
}

// LiEnd closes li HTML tag
func LiEnd(vars *map[string]string, pars ...string) string {
	return fmt.Sprintf(`</li>`)
}

// SetVar assigns the value to the variable
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

// TextHidden returns hidden textarea HTML tag
func TextHidden(vars *map[string]string, pars ...string) (out string) {
	for _, item := range pars {
		out += fmt.Sprintf(`<textarea style="display:none;" id="%s">%s</textarea>`, item, (*vars)[item])
	}
	return
}

// TxID returns the integer value of the transaction
func TxID(vars *map[string]string, pars ...string) string {
	if len(pars) == 0 {
		return `0`
	}
	return converter.Int64ToStr(utils.TypeInt(pars[0]))
}

// LinkPage returns the HTML link to the template page
func LinkPage(vars *map[string]string, pars ...string) string {
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

// BtnEdit returns button HTML tag with an icon
func BtnEdit(vars *map[string]string, pars ...string) string {
	params := ``
	if len(pars) < 2 {
		return ``
	}
	if len(pars) >= 3 {
		params = pars[2]
	}
	return fmt.Sprintf(`<button style="width: 44px;" type="button" class="btn btn-labeled btn-default" onclick="load_template('%s', {%s})"><span class="btn-label"><em class="fa fa-%s"></em></span></button>`,
		converter.Escape(pars[0]), params, converter.Escape(pars[1]))
}

// BlockInfo returns returns a link for popup block
func BlockInfo(vars *map[string]string, pars ...string) string {
	return fmt.Sprintf(`<a href="#" onclick="openBlockDetailPopup('%s')">%[1]s</a>`, pars[0])
}

// Val returns the value of the html control with id identifier
func Val(vars *map[string]string, pars ...string) string {
	return fmt.Sprintf(`$('#%s').val()`, pars[0])
}

// BtnPage returns the button HTML tag with the link to the template page
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
	html := `<button type="button" class="%s" %s onclick="load_template('%s', {%s}, %s )">%s</button>`
	page := pars[0]
	if page[0:4] == `app-` {
		html = `<button type="button" class="%s" %s onclick="load_app('%s', {%s}, %s )">%s</button>`
		page = page[4:]
	}
	return fmt.Sprintf(html,
		class, more, page, params, anchor, pars[1])
}

// BtnContract returns the button for executing of the contract
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
		onsuccess = converter.Escape(pars[5])
		page = converter.Escape(pars[6])
		if len(pars) == 8 {
			pageparam = converter.Escape(pars[7])
		}
	}
	(*vars)["wibtncont"] = `1`
	return fmt.Sprintf(`<button type="button" class=%s data-tool="panel-refresh" onclick="btn_contract(this, '%s', {%s}, '%s', '%s', '%s', {%s})">%s</button>`,
		class, pars[0], params, pars[2], onsuccess, page, pageparam, pars[1])
}

// StateLink returns the value of the variable
func StateLink(vars *map[string]string, pars ...string) string {
	if len(pars) < 2 {
		return ``
	}
	return (*vars)[fmt.Sprintf(`%s_%s`, pars[0], pars[1])]
}

// Table returns table HTML tag with the result of sql query
func Table(vars *map[string]string, pars *map[string]string) string {
	fields := `*`
	order := ``
	where := ``
	limit := ``
	tableClass := ``
	tableMore := ``
	adaptive := ``
	if val, ok := (*pars)[`Order`]; ok {
		order = `order by ` + converter.Escape(val)
	}
	if val, ok := (*pars)[`Class`]; ok {
		tableClass, tableMore = getClass(val)
	}
	if _, ok := (*pars)[`Adaptive`]; ok {
		adaptive = `data-role="table"`
	}
	if val, ok := (*pars)[`Where`]; ok {
		where = `where ` + converter.Escape(val)
	}
	if val, ok := (*pars)[`Limit`]; ok && len(val) > 0 {
		opar := strings.Split(val, `,`)
		if len(opar) == 1 {
			limit = fmt.Sprintf(` limit %d`, converter.StrToInt64(opar[0]))
		} else {
			limit = fmt.Sprintf(` offset %d limit %d`, converter.StrToInt64(opar[0]), converter.StrToInt64(opar[1]))
		}
	}
	if val, ok := (*pars)[`Fields`]; ok {
		fields = converter.Escape(val)
	}
	list, err := sql.DB.GetAll(fmt.Sprintf(`select %s from %s %s %s%s`, fields,
		converter.EscapeName((*pars)[`Table`]), where, order, limit), -1)
	if err != nil {
		return err.Error()
	}
	columns := textproc.Split((*pars)[`Columns`])

	out := ``
	if strings.TrimSpace(tableClass) == `table-responsive` {
		out += `<div class="table-responsive">`
	}
	out += `<table class="table ` + tableClass + `" ` + tableMore + ` ` + adaptive + `><thead>`
	for _, th := range *columns {
		if len(th) < 2 {
			return `incorrect column`
		}
		class := ``
		more := ``
		if len(th) > 2 {
			class, more = getClass(th[2])
			if len(class) > 0 {
				class = fmt.Sprintf(`class="%s"`, class)
			}
		}
		out += fmt.Sprintf(`<th %s %s>`, class, more) + th[0] + `</th>`
		th[1] = strings.TrimSpace(th[1])
		off := strings.Index(th[1], `StateLink`)

		if off >= 0 {
			thname := th[1][off:]
			if strings.IndexByte(thname, ',') > 0 {
				linklist := strings.TrimSpace(thname[strings.IndexByte(thname, '(')+1 : strings.IndexByte(thname, ',')])
				if alist := strings.Split(StateVal(vars, linklist), `,`); len(alist) > 0 {
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
			if value == `NULL` {
				value = ``
			}
			if key != `state_id` {
				(*vars)[key] = converter.StripTags(value)
			}
		}
		for _, th := range *columns {
			if len(th) < 2 {
				return `incorrect column`
			}
			//			val := textproc.Process(th[1], vars)
			val := textproc.Macro(th[1], vars)
			if val == `NULL` {
				val = ``
			}
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

// TxForm returns HTML form for the contract
func TxForm(vars *map[string]string, pars *map[string]string) string {
	return TXForm(vars, pars)
}

// Image returns image HTML tag
func Image(vars *map[string]string, pars ...string) string {
	alt := ``
	class := ``
	more := ``
	if len(pars) > 1 {
		alt = pars[1]
	}
	if len(pars) > 2 {
		class, more = getClass(pars[2])
	}
	rez := " "
	if len(pars[0]) == 0 || (strings.HasPrefix(pars[0], `data:`) || strings.HasSuffix(pars[0], `jpg`) ||
		strings.HasSuffix(pars[0], `png`) || strings.HasSuffix(pars[0], `svg`) || strings.HasSuffix(pars[0], `gif`)) {
		rez = fmt.Sprintf(`<img src="%s" class="%s" %s alt="%s" stylex="display:block;">`, pars[0], class, more, alt)
	}
	return rez
}

// InputMap returns HTML tags for map point
func InputMap(vars *map[string]string, pars ...string) string {
	var coords string
	id := pars[0]
	if len(id) == 0 {
		return ``
	}
	if len(pars) > 1 {
		coords = strings.Replace(pars[1], `<`, `&lt;`, -1)
	}
	(*vars)[`inmappoint`] = `1`
	out := fmt.Sprintf(`<div class="form-group"><label>Map</label><textarea class="form-control inmap" id="%s">%s</textarea></div>`, 
			   id, coords)
	if len(pars) > 2 {
		out += fmt.Sprintf(`<div class="form-group"><label>Address</label><input type="text" class="form-control" 
		        id="%s_address" value="%s"></div>`, id, strings.Replace(pars[2], `<`, `&lt;`, -1))
	}
	return out
}

// InputMapPoly returns HTML tags for polygon map
func InputMapPoly(vars *map[string]string, pars ...string) string {
	var coords string
	id := pars[0]
	if len(id) == 0 {
		return ``
	}
	if len(pars) > 1 {
		coords = strings.Replace(pars[1], `<`, `&lt;`, -1)
	}
	out := fmt.Sprintf(`<div class="form-group"><label>Map</label><textarea class="form-control" id="%s">%s</textarea>
         <button type="button" onClick="openMap('%[1]s');" class="btn btn-primary"><i class="fa fa-map-marker"></i> &nbsp;Add/Edit Coords</button></div>`, id, coords)
	if len(pars) > 2 {
		out += fmt.Sprintf(`<div class="form-group"><label>Address</label><input type="text" class="form-control" 
		        id="%s_address" value="%s"></div>`, id, strings.Replace(pars[2], `<`, `&lt;`, -1))
	}
	return out
}

// ImageInput returns HTML tags for uploading image
func ImageInput(vars *map[string]string, pars ...string) string {
	id := pars[0]
	if len(id) == 0 {
		return ``
	}
	width := 100
	height := 100
	ratio := `1/1`
	if len(pars) > 1 {
		width = converter.StrToInt(pars[1])
	}
	if len(pars) > 2 {
		var w, h int
		if lr := strings.Split(pars[2], `/`); len(lr) == 2 {
			w, h = converter.StrToInt(lr[0]), converter.StrToInt(lr[1])
			height = int(width * w / h)
		} else {
			height = converter.StrToInt(pars[2])
			w, h = width, height
			for _, i := range []int{2, 3, 5, 7} {
				for (w%i) == 0 && (h%i) == 0 {
					w = w / i
					h = h / i
				}
			}
		}
		ratio = fmt.Sprintf(`%d/%d`, w, h)
	}
	return fmt.Sprintf(`<textarea style="display:none" class="form-control" id="%[1]s"></textarea>
			<button type="button" class="btn btn-primary" onClick="openImageEditor('img%[1]s', '%[1]s', '%s', '%d', '%d');">
			<i class="fa fa-file-image-o"></i> &nbsp;Add/Edit Image</button>`, id, ratio, width, height)
}

// StateVal returns par[1]-th value of pars[0] state param
func StateVal(vars *map[string]string, pars ...string) string {
	val, _ := StateParam(converter.StrToInt64((*vars)[`state_id`]), pars[0])
	if len(pars) > 1 {
		ind := converter.StrToInt(pars[1])
		if alist := strings.Split(val, `,`); ind > 0 && len(alist) >= ind {
			val = LangRes(vars, alist[ind-1])
		} else {
			val = ``
		}
	}
	return val
}

// LiTemplate returns li HTML tag with a link to the template page
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

// Navigation returns bread crumb navigation links
func Navigation(vars *map[string]string, pars ...string) string {
	li := make([]string, 0)
	for _, ipar := range pars {
		li = append(li, ipar)
	}
	return textproc.Macro(fmt.Sprintf(`<ol class="breadcrumb"><span class="pull-right">
	<a href='#' onclick="load_template('sys-editPage', {name: '#page#', global:'#global#'} )">Edit</a></span>%s</ol>`,
		strings.Join(li, `&nbsp;/&nbsp;`)), vars)
}

// MarkDown returns processed markdown text
func MarkDown(vars *map[string]string, pars ...string) string {
	return textproc.Macro(string(blackfriday.MarkdownCommon([]byte(pars[0]))), vars)
}

// Title returns a div tag with the title class
func Title(vars *map[string]string, pars ...string) string {
	return fmt.Sprintf(`<div class="content-heading">%s</div>`, pars[0])
}

// PageTitle returns the header of the page panel
func PageTitle(vars *map[string]string, pars ...string) string {
	var row string
	if len((*vars)[`isrow`]) == 0 {
		row = ` row`
		(*vars)[`isrow`] = `closed`
	}
	return fmt.Sprintf(`<div class="panel panel-default" data-sweet-alert><div class="panel-heading"><div class="panel-title">%s</div></div><div class="panel-body%s">`, pars[0], row)
}

// PageEnd closes the page panel
func PageEnd(vars *map[string]string, pars ...string) string {
	if (*vars)[`isrow`] == `closed` {
		(*vars)[`isrow`] = ``
	}
	return `</div></div>`
}

// Form returns the form HTML tag
func Form(vars *map[string]string, pars ...string) string {
	var class string
	if len(pars[0]) > 0 {
		class = fmt.Sprintf(`class="%s"`, pars[0])
	}
	return fmt.Sprintf(`<form role="form" %s>`, class)
}

// FormEnd closes the form HTML tag
func FormEnd(vars *map[string]string, pars ...string) string {
	return `</form>`
}

// Label returns the label HTML tag
func Label(vars *map[string]string, pars ...string) string {
	var class string
	if len(pars) > 1 && len(pars[1]) > 0 {
		class = fmt.Sprintf(`class="%s"`, pars[1])
	}
	text := LangRes(vars, pars[0])
	return fmt.Sprintf(`<label %s>%s</label>`, class, text)
}

// Legend returns the legend HTML tag
func Legend(vars *map[string]string, pars ...string) (out string) {
	return getTag(`legend`, pars...)
}

// GetVar returns the processed value of the variable
func GetVar(vars *map[string]string, pars ...string) (out string) {
	if val, ok := (*vars)[pars[0]]; ok {
		out = textproc.Process(val, vars)
		if out == `NULL` {
			out = textproc.Macro(val, vars)
		}
	}
	return
}

// ValueByID gets a row from table with the specified id and aasigns the values of fields to variables
func ValueByID(vars *map[string]string, pars ...string) string {
	// tablename, id of value, parameters
	if len(pars) < 3 {
		return ``
	}
	value, err := sql.DB.OneRow(`select * from ` + converter.EscapeName(pars[0]) + ` where id='` + converter.Escape(pars[1]) + `'`).String()
	if err != nil {
		return err.Error()
	}
	keys := make(map[string]string)
	src := strings.Split(converter.Escape(pars[2]), `,`)
	if len(pars) == 4 {
		dest := strings.Split(converter.Escape(pars[3]), `,`)
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

// TXButton returns button which calls the contract
func TXButton(vars *map[string]string, pars *map[string]string) string {
	var unique int64
	if uval, ok := (*vars)[`tx_unique`]; ok {
		unique = converter.StrToInt64(uval) + 1
	}
	(*vars)[`tx_unique`] = converter.Int64ToStr(unique)
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
	contract := smart.GetContract(name, uint32(converter.StrToUint64((*vars)[`state_id`])))
	if contract == nil /*|| contract.Block.Info.(*script.ContractInfo).Tx == nil*/ {
		return fmt.Sprintf(`there is not %s contract`, name)
	}
	funcMap := template.FuncMap{
		"sum": func(a, b interface{}) float64 {
			return converter.InterfaceToFloat64(a) + converter.InterfaceToFloat64(b)
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
			onsuccess = converter.Escape(pars[0])
		}
	}

	b := new(bytes.Buffer)
	finfo := TxButtonInfo{TxName: name, Class: class, ClassBtn: classBtn, Name: LangRes(vars, btnName),
		Unique: template.JS((*vars)[`tx_unique`]), OnSuccess: template.JS(onsuccess),
		Fields: make([]TxInfo, 0), AutoClose: (*pars)[`AutoClose`] != `0`,
		Silent: (*pars)[`Silent`] == `1`}

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
					finfo.Fields = append(finfo.Fields, TxInfo{Name: fitem.Name, ID: idname, Value: value, HTMLType: tag})
					continue txlist
				}
			}
			if fitem.Type.String() == script.Decimal {
				var count int
				if ret := regexp.MustCompile(`(?is)digit:(\d+)`).FindStringSubmatch(fitem.Tags); len(ret) == 2 {
					count = converter.StrToInt(ret[1])
				} else {
					count = converter.StrToInt(StateVal(vars, `money_digit`))
				}
				finfo.Fields = append(finfo.Fields, TxInfo{Name: fitem.Name, Value: value, HTMLType: "money",
					ID: idname, Param: converter.IntToStr(count)})
			} else if fitem.Type.String() == `[]interface {}` {
				finfo.Fields = append(finfo.Fields, TxInfo{Name: fitem.Name, Value: value, ID: idname, HTMLType: "array"})
			} else {
				finfo.Fields = append(finfo.Fields, TxInfo{Name: fitem.Name, Value: value, ID: idname, HTMLType: "textinput"})
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
	tblname := converter.EscapeName(tbl[0])
	name = tbl[1]
	id = `id`
	if len(tbl) > 2 {
		id = tbl[2]
	}
	count, err = sql.DB.Single(`select count(*) from ` + tblname).Int64()
	if err != nil {
		return
	}
	if count > 0 && count <= 50 {
		data, err = sql.DB.GetAll(fmt.Sprintf(`select %s, %s from %s order by %s`, id,
			converter.EscapeName(name), tblname, converter.EscapeName(name)), -1)
	}
	return
}

// TXForm returns HTML form for the contract
func TXForm(vars *map[string]string, pars *map[string]string) string {
	var unique int64
	if uval, ok := (*vars)[`tx_unique`]; ok {
		unique = converter.StrToInt64(uval) + 1
	}
	(*vars)[`tx_unique`] = converter.Int64ToStr(unique)
	name := (*pars)[`Contract`]
	//	init := (*pars)[`Init`]
	//fmt.Println(`TXForm Init`, *vars)
	onsuccess := (*pars)[`OnSuccess`]
	contract := smart.GetContract(name, uint32(converter.StrToUint64((*vars)[`state_id`])))
	if contract == nil || contract.Block.Info.(*script.ContractInfo).Tx == nil {
		return fmt.Sprintf(`there is not %s contract or parameters`, name)
	}
	funcMap := template.FuncMap{
		"sum": func(a, b interface{}) float64 {
			return converter.InterfaceToFloat64(a) + converter.InterfaceToFloat64(b)
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
			onsuccess = converter.Escape(pars[0])
		}
	}

	b := new(bytes.Buffer)
	finfo := FormInfo{TxName: name, Unique: template.JS((*vars)[`tx_unique`]), OnSuccess: template.JS(onsuccess),
		Fields: make([]FieldInfo, 0), AutoClose: (*pars)[`AutoClose`] != `0`,
		Silent: (*pars)[`Silent`] == `1`}

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
			finfo.Fields = append(finfo.Fields, FieldInfo{Name: fitem.Name, HTMLType: `hidden`,
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
				finfo.Fields = append(finfo.Fields, FieldInfo{Name: fitem.Name, HTMLType: tag,
					TxType: fitem.Type.String(), Title: title, Value: value})
				continue txlist
			}
		}
		if len(linklist) > 0 {
			sellist := SelList{converter.StrToInt64(value), make(map[int]string)}
			if strings.IndexByte(linklist, '.') >= 0 {
				if data, id, name, err := getSelect(linklist); err != nil {
					return err.Error()
				} else if len(data) > 0 {
					for _, item := range data {
						sellist.List[int(converter.StrToInt64(item[id]))] = converter.StripTags(item[name])
					}
				}
			} else if alist := strings.Split(StateVal(vars, linklist), `,`); len(alist) > 0 {
				for ind, item := range alist {
					sellist.List[ind+1] = LangRes(vars, item)
				}
			}
			finfo.Fields = append(finfo.Fields, FieldInfo{Name: fitem.Name, HTMLType: "select",
				TxType: fitem.Type.String(), Title: title, Value: sellist})
		} else if fitem.Type.String() == script.Decimal {
			var count int
			if ret := regexp.MustCompile(`(?is)digit:(\d+)`).FindStringSubmatch(fitem.Tags); len(ret) == 2 {
				count = converter.StrToInt(ret[1])
			} else {
				count = converter.StrToInt(StateVal(vars, `money_digit`))
			}
			value = Money(vars, value, converter.IntToStr(count))
			finfo.Fields = append(finfo.Fields, FieldInfo{Name: fitem.Name, HTMLType: "money",
				TxType: fitem.Type.String(), Title: title, Value: value,
				Param: converter.IntToStr(count) /*`9{1,20}` + postfix*/})
		} else if fitem.Type.String() == `string` || fitem.Type.String() == `int64` || fitem.Type.String() == `float64` {
			finfo.Fields = append(finfo.Fields, FieldInfo{Name: fitem.Name, HTMLType: "textinput",
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

// IDToAddress converts the number to the wallet address
func IDToAddress(vars *map[string]string, pars ...string) string {
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
	return converter.AddressToString(id)
}

// Ring returns a ring HTML control
func Ring(vars *map[string]string, pars ...string) string {
	count := 0
	size := 18
	if len(pars) > 0 {
		count = int(converter.StrToInt64(pars[0]))
	}
	if len(pars) > 1 {
		size = int(converter.StrToInt64(pars[1]))
	}
	pct := 100
	if len(pars) > 2 {
		pct = int(converter.StrToInt64(pars[2]))
	}
	speed := 1
	if len(pars) > 3 {
		speed = int(converter.StrToInt64(pars[3]))
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
		width = int(converter.StrToInt64(pars[6]))
	}
	thickness := 10
	if len(pars) > 7 {
		thickness = int(converter.StrToInt64(pars[7]))
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

// WiBalance returns a balance widget
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
			   </div></div></div></div>`, converter.NumString(pars[0]), converter.Escape(pars[1]))
}

// WiAccount returns an account widget
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
			</div></div></div>`, converter.Escape(pars[0]))
}

// Source returns HTML control for source code
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

// WiCitizen returns a widget with the information about the citizen
func WiCitizen(vars *map[string]string, pars ...string) string {
	image := `/static/img/apps/ava.png`
	flag := ``
	if len(pars) < 2 {
		return ``
	}
	if len(pars) > 2 && pars[2] != `NULL` && pars[2] != `` && pars[2] != `#my_avatar#` {
		image = pars[2]
	}
	if len(pars) > 3 && len(pars[3]) > 0 {
		flag = fmt.Sprintf(`<img src="%s" alt="Image" class="wd-xs">`, pars[3])
	}
	address := converter.AddressToString(converter.StrToInt64(pars[1]))
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
		</div></div></div></div>`, image, converter.Escape(pars[0]), flag, address, address)
}

// Mult multiplies two float64 values
func Mult(vars *map[string]string, pars ...string) string {
	if len(pars) != 2 {
		return ``
	}
	return converter.Int64ToStr(converter.RoundWithoutPrecision(converter.StrToFloat64(pars[0]) * converter.StrToFloat64(pars[1])))
}

// Date formats the date value
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

// DateTime formats the date/time value
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

// Select returns select HTML tag
func Select(vars *map[string]string, pars ...string) string {
	var (
		class, more string
		value       int64
	)
	list := make([]SelInfo, 0)
	if len(pars) > 1 {
		if strings.IndexByte(pars[1], '.') >= 0 {
			if data, id, name, err := getSelect(pars[1]); err != nil {
				return err.Error()
			} else if len(data) > 0 {
				for _, item := range data {
					list = append(list, SelInfo{ID: converter.StrToInt64(item[id]), Name: converter.StripTags(item[name])})
				}
			}
		} else if alist := strings.Split(StateVal(vars, pars[1]), `,`); len(alist) > 0 {
			for ind, item := range alist {
				list = append(list, SelInfo{ID: int64(ind + 1), Name: LangRes(vars, item)})
			}
		}
	}
	if len(pars) > 2 {
		class, more = getClass(pars[2])
	}
	if len(pars) > 3 {
		value = converter.StrToInt64(pars[3])
	}

	out := fmt.Sprintf(`<select id="%s" class="selectbox form-control %s" %s>`, pars[0], class, more)
	for _, item := range list {
		var selected string
		if item.ID == value {
			selected = `selected`
		}
		out += fmt.Sprintf(`<option value="%d" %s>%s</option>`, item.ID, selected, item.Name)

	}
	return out + `</select>`
}

func mapOut(vars *map[string]string, mapClass string, pars []string) string {
	class := ``
	more := ``
	if len(pars) > 1 {
		class, more = getClass(pars[1])
	}
	(*vars)[mapClass] = `1`
	return fmt.Sprintf(`<div class="%s %s" %s>%s</div>`, mapClass, class, more, pars[0])
}

// Map returns a map widget
func Map(vars *map[string]string, pars ...string) string {
	return mapOut(vars, `wimap`, pars)
}

// MapPoint returns a map widget with a baloon
func MapPoint(vars *map[string]string, pars ...string) string {
	return mapOut(vars, `wimappoint`, pars)
}

// MenuGroup returns a group of the menu items
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
		idname = converter.Escape(pars[1])
	}
	if len(pars) > 2 {
		icon = fmt.Sprintf(`<em class="%s"></em>`, converter.Escape(pars[2]))
	}
	return fmt.Sprintf(`<li id="li%s"><span>%s
     <span>%s</span></span>
	 <ul id="ul%[1]s">`,
		idname, icon, LangRes(vars, converter.Escape(pars[0])))
}

// MenuItem returns a menu item
func MenuItem(vars *map[string]string, pars ...string) string {
	var (
		/*idname,*/ action, page, params, icon string
	)
	/*	if len(pars) > 1 {
		idname = lib.Escape(pars[1])
	}*/
	off := 0
	if len(pars) > 1 {
		action = converter.Escape(pars[1])
	}
	if !strings.HasPrefix(action, `load_`) {
		action = `load_template`
		off = 1
	}
	if len(pars) > 2-off {
		page = converter.Escape(pars[2-off])
	}
	if len(pars) > 3-off {
		params = converter.Escape(pars[3-off])
	}
	if len(pars) > 4-off {
		icon = fmt.Sprintf(`<em class="%s"></em>`, converter.Escape(pars[4-off]))
	}
	return fmt.Sprintf(`<li id="li%s">
		<a href="#" title="%s" onClick="%s('%s',{%s});">
		%s<span>%[2]s</span></a></li>`,
		page, LangRes(vars, converter.Escape(pars[0])), action, page, params, icon)
}

// MenuPage returns a special comment for the menu
func MenuPage(vars *map[string]string, pars ...string) string {
	return fmt.Sprintf(`<!--%s-->`, converter.Escape(pars[0]))
}

// MenuBack returns a special menu link
func MenuBack(vars *map[string]string, pars ...string) string {
	var link string
	if len(pars) > 1 {
		link = fmt.Sprintf(`load_template('%s')`, converter.Escape(pars[1]))
	}
	return fmt.Sprintf(`<!--%s=%s-->`, converter.Escape(pars[0]), link)
}

// MenuEnd closes menu tags
func MenuEnd(vars *map[string]string, pars ...string) string {
	return `</ul></li>`
}

// ChartBar returns bar chart with the information from the database
func ChartBar(vars *map[string]string, pars *map[string]string) string {
	id := fmt.Sprintf(`bar%d`, crypto.RandInt(0, 0xfffffff))
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
		order = `order by ` + converter.Escape(val)
	}
	if val, ok := (*pars)[`Where`]; ok {
		where = `where ` + converter.Escape(val)
	}
	if val, ok := (*pars)[`Limit`]; ok && len(val) > 0 {
		opar := strings.Split(val, `,`)
		if len(opar) == 1 {
			limit = fmt.Sprintf(` limit %d`, converter.StrToInt64(opar[0]))
		} else {
			limit = fmt.Sprintf(` offset %d limit %d`, converter.StrToInt64(opar[0]), converter.StrToInt64(opar[1]))
		}
	}
	list, err := sql.DB.GetAll(fmt.Sprintf(`select %s,%s from %s %s %s%s`, converter.EscapeName(value), converter.EscapeName(label),
		converter.EscapeName((*pars)[`Table`]), where, order, limit), -1)
	if err != nil {
		return err.Error()
	}
	for _, item := range list {
		if item[value] == `NULL` {
			item[value] = ``
		}
		data = append(data, converter.StripTags(item[value]))
		labels = append(labels, `'`+converter.StripTags(item[label])+`'`)
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

// ChartPie returns pie chart with the information from the database
func ChartPie(vars *map[string]string, pars *map[string]string) string {
	id := fmt.Sprintf(`pie%d`, crypto.RandInt(0, 0xfffffff))
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
			order = `order by ` + converter.Escape(val)
		}
		if val, ok := (*pars)[`Where`]; ok {
			where = `where ` + converter.Escape(val)
		}
		if val, ok := (*pars)[`Limit`]; ok && len(val) > 0 {
			opar := strings.Split(val, `,`)
			if len(opar) == 1 {
				limit = fmt.Sprintf(` limit %d`, converter.StrToInt64(opar[0]))
			} else {
				limit = fmt.Sprintf(` offset %d limit %d`, converter.StrToInt64(opar[0]), converter.StrToInt64(opar[1]))
			}
		} else {
			limit = fmt.Sprintf(` limit %d`, len(colors))
		}
		list, err := sql.DB.GetAll(fmt.Sprintf(`select %s,%s from %s %s %s%s`, converter.EscapeName(value), converter.EscapeName(label),
			converter.EscapeName((*pars)[`Table`]), where, order, limit), -1)
		if err != nil {
			return err.Error()
		}
		for ind, item := range list {
			if item[value] == `NULL` {
				item[value] = ``
			}
			color := colors[ind%len(colors)]
			out = append(out, fmt.Sprintf(`{
				value: %s,
				color: '#%s',
				highlight: '#%s',
				label: '%s'
			}`, converter.StripTags(item[value]), color, color, converter.StripTags(item[label])))
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

// ProceedTemplate proceeds html template
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
	//fmt.Println(`PROC`, err, b.String())
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

// CreateHTMLFromTemplate gets the template of the page from the table and proceeds it
func CreateHTMLFromTemplate(page string, citizenID, stateID int64, params *map[string]string) (string, error) {
	var data string
	var err error
	query := `SELECT value FROM "` + converter.Int64ToStr(stateID) + `_pages" WHERE name = ?`
	if (*params)[`global`] == `1` {
		query = `SELECT value FROM global_pages WHERE name = ?`
	}
	if page == `body` && len((*params)[`autobody`]) > 0 {
		data = (*params)[`autobody`]
	} else {
		data, err = sql.DB.Single(query, page).String()
		if err != nil {
			return "", err
		}
	}
	(*params)[`page`] = page
	(*params)[`state_id`] = converter.Int64ToStr(stateID)
	(*params)[`citizen`] = converter.Int64ToStr(citizenID)
	if len(data) > 0 {
		templ := textproc.Process(data, params)
		if (*params)[`isrow`] == `opened` {
			templ += `</div>`
			(*params)[`isrow`] = ``
		}
		templ = language.LangMacro(templ, int(stateID), (*params)[`accept_lang`])
		getHeight := func() int64 {
			height := int64(100)
			if h, ok := (*params)[`hmap`]; ok {
				height = converter.StrToInt64(h)
			}
			return height
		}
		if len((*params)[`wisource`]) > 0 {
			templ += fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
			var editor = ace.edit("textEditor");
	var ContractMode = ace.require("ace/mode/c_cpp").Mode;
	ace.require("ace/ext/language_tools");
	$(".textEditor code").html(editor.getValue());
	$("#%s").val(editor.getValue());
	editor.setTheme("ace/theme/chrome");
    editor.session.setMode(new ContractMode());
	editor.setShowPrintMargin(false);
	editor.getSession().setTabSize(4);
	editor.getSession().setUseWrapMode(true);
	editor.getSession().on('change', function(e) {
		$(".textEditor code").html(editor.getValue());
		$("#%s").val(editor.getValue());
		editor.resize();
	});
	editor.setOptions({
		enableBasicAutocompletion: true,
		enableSnippets: true,
		enableLiveAutocompletion: true
	});
			</script>`, (*params)[`wisource`], (*params)[`wisource`])
		}
		if (*params)[`wimoney`] == `1` {
			templ += fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
				$(".inputmask").inputmask({'autoUnmask': true});</script>`)
		}
		if (*params)[`widate`] == `1` {
			templ += fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
						$(document).ready(function() {
							$.datetimepicker.setLocale('en');
							$(".datetimepicker").datetimepicker();
						})
				</script>`)
		}
		if (*params)[`wiaddress`] == `1` {
			templ += fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
				$(".address").prop("autocomplete", "off").inputmask({mask: "9999-9999-9999-9999-9999", autoUnmask: true }).focus();
	$(".address").typeahead({
		minLength: 1,
		items: 10,
		source: function (query, process) {
			return $.get('ajax?json=ajax_addresses', { 'address': query }, function (data) {
				return process(data.address);
			});
		}
	}).focus();</script>`)
		}
		if (*params)[`wimap`] == `1` {
			templ += fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
			miniMap("wimap", "100%%", "%dpx");</script>`, getHeight())
		}
		if (*params)[`wicitizen`] == `1` {
			templ += fmt.Sprintf(`<script language="JavaScript" type="text/javascript">(function($, window, document){
'use strict';
  var Selector = '[data-notify]',
      autoloadSelector = '[data-onload]',
      doc = $(document);

  $(function() {
    $(Selector).each(function(){
      var $this  = $(this),
          onload = $this.data('onload');
      if(onload !== undefined) {
        setTimeout(function(){
          notifyNow($this);
        }, 800);
      }
      $this.on('click', function (e) {
        e.preventDefault();
        notifyNow($this);
      });
    });
  });
  function notifyNow($element) {
      var message = $element.data('message'),
          options = $element.data('options');
 	 if(!message)
        $.error('Notify: No message specified');
      $.notify(message, options || {});
  }
}(jQuery, window, document));</script>`)
		}
		if (*params)[`wimappoint`] == `1` {
			templ += fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
			userLocation("wimappoint", "100%%", "%dpx");</script>`, getHeight())
		}
		if (*params)[`inmappoint`] == `1` {
			templ += fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
			userLocation("inmap", "100%%", "%dpx");</script>`, getHeight())
		}
		if (*params)[`wibtncont`] == `1` {
			var unique int64
			if uval, ok := (*params)[`tx_unique`]; ok {
				unique = converter.StrToInt64(uval) + 1
			}
			(*params)[`tx_unique`] = converter.Int64ToStr(unique)
			funcMap := template.FuncMap{
				"sum": func(a, b interface{}) float64 {
					return converter.InterfaceToFloat64(a) + converter.InterfaceToFloat64(b)
				},
				"noescape": func(s string) template.HTML {
					return template.HTML(s)
				},
			}
			data, err := static.Asset("static/tx_btncont.html")
			if err != nil {
				return ``, err
			}
			sign, err := static.Asset("static/signatures_new.html")
			if err != nil {
				return ``, err
			}

			t := template.New("template").Funcs(funcMap)
			if t, err = t.Parse(string(data)); err != nil {
				return ``, err
			}
			t = template.Must(t.Parse(string(sign)))
			b := new(bytes.Buffer)

			finfo := TxBtnCont{Unique: template.JS((*params)[`tx_unique`])}
			if err = t.Execute(b, finfo); err != nil {
				return ``, err
			}
			templ += b.String()
		}
		return ProceedTemplate(`page_template`, &PageTpl{Page: page, Template: templ})
	}
	return ``, nil
}
