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
	"crypto/md5"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/config/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/language"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/smart"

	log "github.com/sirupsen/logrus"
)

var (
	funcs = make(map[string]tplFunc)
	tails = make(map[string]forTails)
	modes = [][]rune{{'(', ')'}, {'{', '}'}}
)

func init() {
	funcs[`Lower`] = tplFunc{lowerTag, defaultTag, `lower`, `Text`}
	funcs[`AddToolButton`] = tplFunc{defaultTag, defaultTag, `addtoolbutton`, `Title,Icon,Page,PageParams`}
	funcs[`Address`] = tplFunc{addressTag, defaultTag, `address`, `Wallet`}
	funcs[`Calculate`] = tplFunc{calculateTag, defaultTag, `calculate`, `Exp,Type,Prec`}
	funcs[`CmpTime`] = tplFunc{cmpTimeTag, defaultTag, `cmptime`, `Time1,Time2`}
	funcs[`Code`] = tplFunc{defaultTag, defaultTag, `code`, `Text`}
	funcs[`DateTime`] = tplFunc{dateTimeTag, defaultTag, `datetime`, `DateTime,Format`}
	funcs[`EcosysParam`] = tplFunc{ecosysparTag, defaultTag, `ecosyspar`, `Name,Index,Source`}
	funcs[`Em`] = tplFunc{defaultTag, defaultTag, `em`, `Body,Class`}
	funcs[`GetVar`] = tplFunc{getvarTag, defaultTag, `getvar`, `Name`}
	funcs[`ImageInput`] = tplFunc{defaultTag, defaultTag, `imageinput`, `Name,Width,Ratio,Format`}
	funcs[`InputErr`] = tplFunc{defaultTag, defaultTag, `inputerr`, `*`}
	funcs[`LangRes`] = tplFunc{langresTag, defaultTag, `langres`, `Name,Lang`}
	funcs[`MenuGroup`] = tplFunc{menugroupTag, defaultTag, `menugroup`, `Title,Body,Icon`}
	funcs[`MenuItem`] = tplFunc{defaultTag, defaultTag, `menuitem`, `Title,Page,PageParams,Icon,Vde`}
	funcs[`Now`] = tplFunc{nowTag, defaultTag, `now`, `Format,Interval`}
	funcs[`SetTitle`] = tplFunc{defaultTag, defaultTag, `settitle`, `Title`}
	funcs[`SetVar`] = tplFunc{setvarTag, defaultTag, `setvar`, `Name,Value`}
	funcs[`Strong`] = tplFunc{defaultTag, defaultTag, `strong`, `Body,Class`}
	funcs[`SysParam`] = tplFunc{sysparTag, defaultTag, `syspar`, `Name`}
	funcs[`Button`] = tplFunc{buttonTag, buttonTag, `button`, `Body,Page,Class,Contract,Params,PageParams`}
	funcs[`Div`] = tplFunc{defaultTailTag, defaultTailTag, `div`, `Class,Body`}
	funcs[`ForList`] = tplFunc{forlistTag, defaultTag, `forlist`, `Source,Data,Index`}
	funcs[`Form`] = tplFunc{defaultTailTag, defaultTailTag, `form`, `Class,Body`}
	funcs[`If`] = tplFunc{ifTag, ifFull, `if`, `Condition,Body`}
	funcs[`Image`] = tplFunc{defaultTailTag, defaultTailTag, `image`, `Src,Alt,Class`}
	funcs[`Include`] = tplFunc{includeTag, defaultTag, `include`, `Name`}
	funcs[`Input`] = tplFunc{defaultTailTag, defaultTailTag, `input`, `Name,Class,Placeholder,Type,@Value,Disabled`}
	funcs[`Label`] = tplFunc{defaultTailTag, defaultTailTag, `label`, `Body,Class,For`}
	funcs[`LinkPage`] = tplFunc{defaultTailTag, defaultTailTag, `linkpage`, `Body,Page,Class,PageParams`}
	funcs[`Data`] = tplFunc{dataTag, defaultTailTag, `data`, `Source,Columns,Data`}
	funcs[`DBFind`] = tplFunc{dbfindTag, defaultTailTag, `dbfind`, `Name,Source`}
	funcs[`And`] = tplFunc{andTag, defaultTag, `and`, `*`}
	funcs[`Or`] = tplFunc{orTag, defaultTag, `or`, `*`}
	funcs[`P`] = tplFunc{defaultTailTag, defaultTailTag, `p`, `Body,Class`}
	funcs[`RadioGroup`] = tplFunc{defaultTailTag, defaultTailTag, `radiogroup`, `Name,Source,NameColumn,ValueColumn,Value,Class`}
	funcs[`Span`] = tplFunc{defaultTailTag, defaultTailTag, `span`, `Body,Class`}
	funcs[`Table`] = tplFunc{tableTag, defaultTailTag, `table`, `Source,Columns`}
	funcs[`Select`] = tplFunc{defaultTailTag, defaultTailTag, `select`, `Name,Source,NameColumn,ValueColumn,Value,Class`}
	funcs[`Chart`] = tplFunc{chartTag, defaultTailTag, `chart`, `Type,Source,FieldLabel,FieldValue,Colors`}
	funcs[`InputMap`] = tplFunc{defaultTailTag, defaultTailTag, "inputMap", "Name,@Value,Type,MapType"}
	funcs[`Map`] = tplFunc{defaultTag, defaultTag, "map", "@Value,MapType,Hmap"}

	tails[`button`] = forTails{map[string]tailInfo{
		`Alert`: {tplFunc{alertTag, defaultTailFull, `alert`, `Text,ConfirmButton,CancelButton,Icon`}, true},
		`Style`: {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
	}}
	tails[`div`] = forTails{map[string]tailInfo{
		`Style`: {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
	}}
	tails[`form`] = forTails{map[string]tailInfo{
		`Style`: {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
	}}
	tails[`if`] = forTails{map[string]tailInfo{
		`Else`:   {tplFunc{elseTag, elseFull, `else`, `Body`}, true},
		`ElseIf`: {tplFunc{elseifTag, elseifFull, `elseif`, `Condition,Body`}, false},
	}}
	tails[`image`] = forTails{map[string]tailInfo{
		`Style`: {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
	}}
	tails[`input`] = forTails{map[string]tailInfo{
		`Validate`: {tplFunc{validateTag, validateFull, `validate`, `*`}, false},
		`Style`:    {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
	}}
	tails[`label`] = forTails{map[string]tailInfo{
		`Style`: {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
	}}
	tails[`linkpage`] = forTails{map[string]tailInfo{
		`Style`: {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
	}}
	tails[`data`] = forTails{map[string]tailInfo{
		`Custom`: {tplFunc{customTag, defaultTailFull, `custom`, `Column,Body`}, false},
	}}
	tails[`dbfind`] = forTails{map[string]tailInfo{
		`Columns`:   {tplFunc{tailTag, defaultTailFull, `columns`, `Columns`}, false},
		`Where`:     {tplFunc{tailTag, defaultTailFull, `where`, `Where`}, false},
		`WhereId`:   {tplFunc{tailTag, defaultTailFull, `whereid`, `WhereId`}, false},
		`Order`:     {tplFunc{tailTag, defaultTailFull, `order`, `Order`}, false},
		`Limit`:     {tplFunc{tailTag, defaultTailFull, `limit`, `Limit`}, false},
		`Offset`:    {tplFunc{tailTag, defaultTailFull, `offset`, `Offset`}, false},
		`Ecosystem`: {tplFunc{tailTag, defaultTailFull, `ecosystem`, `Ecosystem`}, false},
		`Custom`:    {tplFunc{customTag, defaultTailFull, `custom`, `Column,Body`}, false},
		`Vars`:      {tplFunc{tailTag, defaultTailFull, `vars`, `Prefix`}, false},
	}}
	tails[`p`] = forTails{map[string]tailInfo{
		`Style`: {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
	}}
	tails[`radiogroup`] = forTails{map[string]tailInfo{
		`Validate`: {tplFunc{validateTag, validateFull, `validate`, `*`}, false},
		`Style`:    {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
	}}
	tails[`span`] = forTails{map[string]tailInfo{
		`Style`: {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
	}}
	tails[`table`] = forTails{map[string]tailInfo{
		`Style`: {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
	}}
	tails[`select`] = forTails{map[string]tailInfo{
		`Validate`: {tplFunc{validateTag, validateFull, `validate`, `*`}, false},
		`Style`:    {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
	}}
	tails[`inputMap`] = forTails{map[string]tailInfo{
		`Validate`: {tplFunc{validateTag, validateFull, `validate`, `*`}, false},
	}}
}

func defaultTag(par parFunc) string {
	setAllAttr(par)
	par.Owner.Children = append(par.Owner.Children, par.Node)
	return ``
}

func lowerTag(par parFunc) string {
	return strings.ToLower((*par.Pars)[`Text`])
}

func menugroupTag(par parFunc) string {
	setAllAttr(par)
	name := (*par.Pars)[`Title`]
	if par.RawPars != nil {
		if v, ok := (*par.RawPars)[`Title`]; ok {
			name = v
		}
	}
	par.Node.Attr[`name`] = name
	par.Owner.Children = append(par.Owner.Children, par.Node)
	return ``
}

func forlistTag(par parFunc) (ret string) {
	var (
		name, indexName string
	)
	setAllAttr(par)
	if len((*par.Pars)[`Source`]) > 0 {
		name = par.Node.Attr[`source`].(string)
	}
	if len((*par.Pars)[`Index`]) > 0 {
		indexName = par.Node.Attr[`index`].(string)
	} else {
		indexName = name + `_index`
	}
	if len(name) == 0 || par.Workspace.Sources == nil {
		return
	}
	source := (*par.Workspace.Sources)[name]
	if source.Data == nil {
		return
	}
	root := node{}
	keys := make(map[string]bool)
	for key := range *par.Workspace.Vars {
		keys[key] = true
	}
	for index, item := range *source.Data {
		vals := map[string]string{indexName: converter.IntToStr(index + 1)}
		for i, icol := range *source.Columns {
			vals[icol] = item[i]
		}
		if index > 0 {
			for key := range *par.Workspace.Vars {
				if !keys[key] {
					delete(*par.Workspace.Vars, key)
				}
			}
		}
		body := replace((*par.Pars)[`Data`], 0, &vals)
		process(body, &root, par.Workspace)
	}
	par.Node.Children = root.Children
	par.Owner.Children = append(par.Owner.Children, par.Node)
	return
}

func addressTag(par parFunc) string {
	idval := (*par.Pars)[`Wallet`]
	if len(idval) == 0 {
		idval = (*par.Workspace.Vars)[`key_id`]
	}
	id, _ := strconv.ParseInt(idval, 10, 64)
	if id == 0 {
		return `unknown address`
	}
	return converter.AddressToString(id)
}

func calculateTag(par parFunc) string {
	return calculate(macro((*par.Pars)[`Exp`], par.Workspace.Vars), (*par.Pars)[`Type`],
		converter.StrToInt((*par.Pars)[`Prec`]))
}

func ecosysparTag(par parFunc) string {
	if len((*par.Pars)[`Name`]) == 0 {
		return ``
	}
	prefix := (*par.Workspace.Vars)[`ecosystem_id`]
	state := converter.StrToInt(prefix)
	if par.Workspace.SmartContract.VDE {
		prefix += `_vde`
	}
	sp := &model.StateParameter{}
	sp.SetTablePrefix(prefix)
	_, err := sp.Get(nil, (*par.Pars)[`Name`])
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting ecosystem param")
		return err.Error()
	}
	val := sp.Value
	if len((*par.Pars)[`Source`]) > 0 {
		data := make([][]string, 0)
		cols := []string{`id`, `name`}
		types := []string{`text`, `text`}
		for key, item := range strings.Split(val, `,`) {
			item, _ = language.LangText(item, state, (*par.Workspace.Vars)[`lang`],
				par.Workspace.SmartContract.VDE)
			data = append(data, []string{converter.IntToStr(key + 1), item})
		}
		node := node{Tag: `data`, Attr: map[string]interface{}{`columns`: &cols, `types`: &types,
			`data`: &data, `source`: (*par.Pars)[`Source`]}}
		par.Owner.Children = append(par.Owner.Children, &node)
		return ``
	}
	if len((*par.Pars)[`Index`]) > 0 {
		ind := converter.StrToInt((*par.Pars)[`Index`])
		if alist := strings.Split(val, `,`); ind > 0 && len(alist) >= ind {
			val, _ = language.LangText(alist[ind-1], state, (*par.Workspace.Vars)[`lang`],
				par.Workspace.SmartContract.VDE)
		} else {
			val = ``
		}
	}
	return val
}

func langresTag(par parFunc) string {
	lang := (*par.Pars)[`Lang`]
	if len(lang) == 0 {
		lang = (*par.Workspace.Vars)[`lang`]
	}
	ret, _ := language.LangText((*par.Pars)[`Name`], int(converter.StrToInt64((*par.Workspace.Vars)[`ecosystem_id`])),
		lang, par.Workspace.SmartContract.VDE)
	return ret
}

func sysparTag(par parFunc) (ret string) {
	if len((*par.Pars)[`Name`]) > 0 {
		ret = syspar.SysString((*par.Pars)[`Name`])
	}
	return
}

// Now returns the current time of postgresql
func nowTag(par parFunc) string {
	var (
		cut   int
		query string
	)
	interval := (*par.Pars)[`Interval`]
	format := (*par.Pars)[`Format`]
	if len(interval) > 0 {
		if interval[0] != '-' && interval[0] != '+' {
			interval = `+` + interval
		}
		interval = fmt.Sprintf(` %s interval '%s'`, interval[:1], strings.TrimSpace(interval[1:]))
	}
	if format == `` {
		query = `select round(extract(epoch from now()` + interval + `))::integer`
		cut = 10
	} else {
		query = `select now()` + interval
		switch format {
		case `datetime`:
			cut = 19
		default:
			if strings.Index(format, `HH`) >= 0 && strings.Index(format, `HH24`) < 0 {
				format = strings.Replace(format, `HH`, `HH24`, -1)
			}
			query = fmt.Sprintf(`select to_char(now()%s, '%s')`, interval, format)
		}
	}
	ret, err := model.Single(query).String()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting single from DB")
		return err.Error()
	}
	if cut > 0 {
		ret = strings.Replace(ret[:cut], `T`, ` `, -1)
	}
	return ret
}

func andTag(par parFunc) string {
	count := len(*par.Pars)
	for i := 0; i < count; i++ {
		if !ifValue((*par.Pars)[strconv.Itoa(i)], par.Workspace) {
			return `0`
		}
	}
	return `1`
}

func orTag(par parFunc) string {
	count := len(*par.Pars)
	for i := 0; i < count; i++ {
		if ifValue((*par.Pars)[strconv.Itoa(i)], par.Workspace) {
			return `1`
		}
	}
	return `0`
}

func alertTag(par parFunc) string {
	setAllAttr(par)
	par.Owner.Attr[`alert`] = par.Node.Attr
	return ``
}

func defaultTailFull(par parFunc) string {
	setAllAttr(par)
	par.Owner.Tail = append(par.Owner.Tail, par.Node)
	return ``
}

func dataTag(par parFunc) string {
	setAllAttr(par)
	defaultTail(par, `data`)

	data := make([][]string, 0)
	cols := strings.Split((*par.Pars)[`Columns`], `,`)
	types := make([]string, len(cols))
	for i := 0; i < len(types); i++ {
		types[i] = `text`
	}

	list, err := csv.NewReader(strings.NewReader((*par.Pars)[`Data`])).ReadAll()
	if err != nil {
		par.Node.Attr[`error`] = err.Error()
	}
	lencol := 0
	defcol := 0
	for _, item := range list {
		if lencol == 0 {
			defcol = len(cols)
			if par.Node.Attr[`customs`] != nil {
				for _, v := range par.Node.Attr[`customs`].([]string) {
					cols = append(cols, v)
					types = append(types, `tags`)
				}
			}
			lencol = len(cols)
		}
		row := make([]string, lencol)
		vals := make(map[string]string)
		for i, icol := range cols {
			var ival string
			if i < defcol {
				ival = strings.TrimSpace(item[i])
				if strings.IndexByte(ival, '<') >= 0 {
					ival = html.EscapeString(ival)
				}
				vals[icol] = ival
			} else {
				body := replace(par.Node.Attr[`custombody`].([]string)[i-defcol], 0, &vals)
				root := node{}
				process(body, &root, par.Workspace)
				out, err := json.Marshal(root.Children)
				if err == nil {
					ival = string(out)
				} else {
					log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling custombody to JSON")
				}
			}
			row[i] = ival
		}
		data = append(data, row)
	}
	setAllAttr(par)
	delete(par.Node.Attr, `customs`)
	delete(par.Node.Attr, `custombody`)
	par.Node.Attr[`columns`] = &cols
	par.Node.Attr[`types`] = &types
	par.Node.Attr[`data`] = &data
	newSource(par)
	par.Owner.Children = append(par.Owner.Children, par.Node)
	return ``
}

func dbfindTag(par parFunc) string {
	var (
		fields string
		state  int64
		err    error
		perm   map[string]string
	)
	if len((*par.Pars)[`Name`]) == 0 {
		return ``
	}
	defaultTail(par, `dbfind`)
	prefix := ``
	where := ``
	order := ``
	limit := 25
	if par.Node.Attr[`columns`] != nil {
		fields = converter.Escape(par.Node.Attr[`columns`].(string))
	}
	if len(fields) == 0 {
		fields = `*`
	}
	fields = strings.ToLower(fields)
	if par.Node.Attr[`where`] != nil {
		where = ` where ` + converter.Escape(par.Node.Attr[`where`].(string))
		where = regexp.MustCompile(`->([\w\d_]+)`).ReplaceAllString(where, "->>'$1'")
	}
	if par.Node.Attr[`whereid`] != nil {
		where = fmt.Sprintf(` where id='%d'`, converter.StrToInt64(par.Node.Attr[`whereid`].(string)))
	}
	if par.Node.Attr[`order`] != nil {
		order = ` order by ` + converter.EscapeName(par.Node.Attr[`order`].(string))
	}
	if par.Node.Attr[`limit`] != nil {
		limit = converter.StrToInt(par.Node.Attr[`limit`].(string))
	}
	if limit > 250 {
		limit = 250
	}
	if par.Node.Attr[`prefix`] != nil {
		prefix = par.Node.Attr[`prefix`].(string)
		limit = 1
	}
	if par.Node.Attr[`ecosystem`] != nil {
		state = converter.StrToInt64(par.Node.Attr[`ecosystem`].(string))
	} else {
		state = converter.StrToInt64((*par.Workspace.Vars)[`ecosystem_id`])
	}
	sc := par.Workspace.SmartContract
	tblname := smart.GetTableName(sc, strings.Trim(converter.EscapeName((*par.Pars)[`Name`]), `"`), state)
	if sc.VDE && *conf.CheckReadAccess {
		perm, err = sc.AccessTablePerm(tblname, `read`)
		cols := strings.Split(fields, `,`)
		if err != nil || sc.AccessColumns(tblname, &cols, false) != nil {
			return `Access denied`
		}
		fields = strings.Join(cols, `,`)
	}
	if fields != `*` && !strings.Contains(fields, `id`) {
		fields += `, id`
	}
	fields = smart.PrepareColumns(fields)

	list, err := model.GetAll(`select `+fields+` from "`+tblname+`"`+where+order, limit)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting all from db")
		return err.Error()
	}
	data := make([][]string, 0)
	cols := make([]string, 0)
	types := make([]string, 0)
	lencol := 0
	defcol := 0
	for _, item := range list {
		if lencol == 0 {
			for key := range item {
				cols = append(cols, key)
				types = append(types, `text`)
			}
			defcol = len(cols)
			if par.Node.Attr[`customs`] != nil {
				for _, v := range par.Node.Attr[`customs`].([]string) {
					cols = append(cols, v)
					types = append(types, `tags`)
				}
			}
			lencol = len(cols)
		}
		row := make([]string, lencol)
		for i, icol := range cols {
			var ival string
			if i < defcol {
				ival = item[icol]
				if strings.IndexByte(ival, '<') >= 0 {
					ival = html.EscapeString(ival)
				}
				if ival == `NULL` {
					ival = ``
				}
				if strings.HasPrefix(ival, `data:image/`) {
					ival = fmt.Sprintf(`/data/%s/%s/%s/%x`, strings.Trim(tblname, `"`),
						item[`id`], icol, md5.Sum([]byte(ival)))
					item[icol] = ival
				}
			} else {
				body := replace(par.Node.Attr[`custombody`].([]string)[i-defcol], 0, &item)
				root := node{}
				process(body, &root, par.Workspace)
				out, err := json.Marshal(root.Children)
				if err == nil {
					ival = string(out)
				} else {
					log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling root children to JSON")
				}
			}
			if par.Node.Attr[`prefix`] != nil {
				(*par.Workspace.Vars)[prefix+`_`+strings.Replace(icol, `.`, `_`, 1)] = ival
			}
			row[i] = ival
		}
		data = append(data, row)
	}
	if sc.VDE && perm != nil && len(perm[`filter`]) > 0 {
		result := make([]interface{}, len(data))
		for i, item := range data {
			row := make(map[string]string)
			for j, col := range cols {
				row[col] = item[j]
			}
			result[i] = reflect.ValueOf(row).Interface()
		}
		fltResult, err := smart.VMEvalIf(sc.VM, perm[`filter`], uint32(sc.TxSmart.EcosystemID),
			&map[string]interface{}{
				`data`:         result,
				`ecosystem_id`: sc.TxSmart.EcosystemID,
				`key_id`:       sc.TxSmart.KeyID, `sc`: sc,
				`block_time`: 0, `time`: sc.TxSmart.Time})
		if err != nil || !fltResult {
			return `Access denied`
		}
		for i := range data {
			for j, col := range cols {
				data[i][j] = result[i].(map[string]string)[col]
			}
		}
	}
	setAllAttr(par)
	delete(par.Node.Attr, `customs`)
	delete(par.Node.Attr, `custombody`)
	delete(par.Node.Attr, `prefix`)
	par.Node.Attr[`columns`] = &cols
	par.Node.Attr[`types`] = &types
	par.Node.Attr[`data`] = &data
	newSource(par)
	par.Owner.Children = append(par.Owner.Children, par.Node)
	return ``
}

func customTag(par parFunc) string {
	setAllAttr(par)
	if par.Owner.Attr[`customs`] == nil {
		par.Owner.Attr[`customs`] = make([]string, 0)
		par.Owner.Attr[`custombody`] = make([]string, 0)
	}
	par.Owner.Attr[`customs`] = append(par.Owner.Attr[`customs`].([]string), par.Node.Attr[`column`].(string))
	par.Owner.Attr[`custombody`] = append(par.Owner.Attr[`custombody`].([]string), (*par.Pars)[`Body`])
	return ``
}

func tailTag(par parFunc) string {
	setAllAttr(par)
	for key, v := range par.Node.Attr {
		par.Owner.Attr[key] = v
	}
	return ``
}

func includeTag(par parFunc) string {
	if len((*par.Pars)[`Name`]) >= 0 && len((*par.Workspace.Vars)[`_include`]) < 5 {
		pattern, err := model.Single(`select value from "`+(*par.Workspace.Vars)[`ecosystem_id`]+`_blocks" where name=?`, (*par.Pars)[`Name`]).String()
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting block by name")
			return err.Error()
		}
		if len(pattern) > 0 {
			root := node{}
			(*par.Workspace.Vars)[`_include`] += `1`
			process(pattern, &root, par.Workspace)
			(*par.Workspace.Vars)[`_include`] = (*par.Workspace.Vars)[`_include`][:len((*par.Workspace.Vars)[`_include`])-1]
			for _, item := range root.Children {
				par.Owner.Children = append(par.Owner.Children, item)
			}
		}
	}
	return ``
}

func setvarTag(par parFunc) string {
	if len((*par.Pars)[`Name`]) > 0 {
		if strings.ContainsAny((*par.Pars)[`Value`], `({`) {
			(*par.Pars)[`Value`] = processToText(par, (*par.Pars)[`Value`])
		}
		(*par.Workspace.Vars)[(*par.Pars)[`Name`]] = (*par.Pars)[`Value`]
	}
	return ``
}

func getvarTag(par parFunc) string {
	if len((*par.Pars)[`Name`]) > 0 {
		return macro((*par.Workspace.Vars)[(*par.Pars)[`Name`]], par.Workspace.Vars)
	}
	return ``
}

func tableTag(par parFunc) string {
	defaultTag(par)
	defaultTail(par, `table`)
	if len((*par.Pars)[`Columns`]) > 0 {
		imap := make([]map[string]string, 0)
		for _, v := range strings.Split((*par.Pars)[`Columns`], `,`) {
			v = strings.TrimSpace(v)
			if off := strings.IndexByte(v, '='); off == -1 {
				imap = append(imap, map[string]string{`Title`: v, `Name`: v})
			} else {
				imap = append(imap, map[string]string{`Title`: strings.TrimSpace(v[:off]), `Name`: strings.TrimSpace(v[off+1:])})
			}
		}
		if len(imap) > 0 {
			par.Node.Attr[`columns`] = imap
		}
	}
	return ``
}

func validateTag(par parFunc) string {
	setAllAttr(par)
	par.Owner.Attr[`validate`] = par.Node.Attr
	return ``
}

func validateFull(par parFunc) string {
	setAllAttr(par)
	par.Owner.Tail = append(par.Owner.Tail, par.Node)
	return ``
}

func defaultTail(par parFunc, tag string) {
	if par.Tails != nil {
		for _, v := range *par.Tails {
			name := (*v)[len(*v)-1]
			curFunc := tails[tag].Tails[string(name)].tplFunc
			pars := (*v)[:len(*v)-1]
			callFunc(&curFunc, par.Node, par.Workspace, &pars, nil)
		}
	}
}

func defaultTailTag(par parFunc) string {
	defaultTag(par)
	defaultTail(par, par.Node.Tag)
	return ``
}

func buttonTag(par parFunc) string {
	defaultTag(par)
	defaultTail(par, `button`)
	return ``
}

func ifTag(par parFunc) string {
	cond := ifValue((*par.Pars)[`Condition`], par.Workspace)
	if cond {
		process((*par.Pars)[`Body`], par.Node, par.Workspace)
		for _, item := range par.Node.Children {
			par.Owner.Children = append(par.Owner.Children, item)
		}
	}
	if !cond && par.Tails != nil {
		for _, v := range *par.Tails {
			name := (*v)[len(*v)-1]
			curFunc := tails[`if`].Tails[string(name)].tplFunc
			pars := (*v)[:len(*v)-1]
			callFunc(&curFunc, par.Owner, par.Workspace, &pars, nil)
			if (*par.Workspace.Vars)[`_cond`] == `1` {
				(*par.Workspace.Vars)[`_cond`] = `0`
				break
			}
		}
	}
	return ``
}

func ifFull(par parFunc) string {
	setAttr(par, `Condition`)
	par.Owner.Children = append(par.Owner.Children, par.Node)
	if par.Tails != nil {
		for _, v := range *par.Tails {
			name := (*v)[len(*v)-1]
			curFunc := tails[`if`].Tails[string(name)].tplFunc
			pars := (*v)[:len(*v)-1]
			callFunc(&curFunc, par.Node, par.Workspace, &pars, nil)
		}
	}
	return ``
}

func elseifTag(par parFunc) string {
	cond := ifValue((*par.Pars)[`Condition`], par.Workspace)
	if cond {
		for _, item := range par.Node.Children {
			par.Owner.Children = append(par.Owner.Children, item)
		}
		(*par.Workspace.Vars)[`_cond`] = `1`
	}
	return ``
}

func elseifFull(par parFunc) string {
	setAttr(par, `Condition`)
	par.Owner.Tail = append(par.Owner.Tail, par.Node)
	return ``
}

func elseTag(par parFunc) string {
	for _, item := range par.Node.Children {
		par.Owner.Children = append(par.Owner.Children, item)
	}
	return ``
}

func elseFull(par parFunc) string {
	par.Owner.Tail = append(par.Owner.Tail, par.Node)
	return ``
}

func dateTimeTag(par parFunc) string {
	datetime := macro((*par.Pars)[`DateTime`], par.Workspace.Vars)
	if len(datetime) == 0 || datetime[0] < '0' || datetime[0] > '9' {
		return ``
	}
	defTime := `1970-01-01T00:00:00`
	lenTime := len(datetime)
	if lenTime < len(defTime) {
		datetime += defTime[lenTime:]
	}
	itime, err := time.Parse(`2006-01-02T15:04:05`, strings.Replace(datetime[:19], ` `, `T`, -1))
	if err != nil {
		return err.Error()
	}
	format := (*par.Pars)[`Format`]
	if len(format) == 0 {
		format, _ = language.LangText(`timeformat`, converter.StrToInt((*par.Workspace.Vars)[`ecosystem_id`]),
			(*par.Workspace.Vars)[`lang`], par.Workspace.SmartContract.VDE)
		if format == `timeformat` {
			format = `2006-01-02 15:04:05`
		}
	} else {
		format = macro(format, par.Workspace.Vars)
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

func cmpTimeTag(par parFunc) string {
	prepare := func(val string) string {
		val = strings.Replace(macro(val, par.Workspace.Vars), `T`, ` `, -1)
		if len(val) > 19 {
			val = val[:19]
		}
		return val
	}
	left := prepare((*par.Pars)[`Time1`])
	right := prepare((*par.Pars)[`Time2`])
	if left == right {
		return `0`
	}
	if left < right {
		return `-1`
	}
	return `1`
}

func chartTag(par parFunc) string {
	defaultTag(par)
	defaultTail(par, "chart")

	if len((*par.Pars)["Colors"]) > 0 {
		colors := strings.Split((*par.Pars)["Colors"], ",")
		for i, v := range colors {
			colors[i] = strings.TrimSpace(v)
		}
		par.Node.Attr["colors"] = colors
	}

	return ""
}
