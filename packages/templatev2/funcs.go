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

package templatev2

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/language"
	"github.com/AplaProject/go-apla/packages/model"
)

var (
	funcs = map[string]tplFunc{
		`Address`:    {addressTag, defaultTag, `address`, `Wallet`},
		`Em`:         {defaultTag, defaultTag, `em`, `Body,Class`},
		`GetVar`:     {getvarTag, defaultTag, `getvar`, `Name`},
		`ImageInput`: {defaultTag, defaultTag, `imageinput`, `Name,Width,Ratio`},
		`InputErr`:   {defaultTag, defaultTag, `inputerr`, `*`},
		`LangRes`:    {langresTag, defaultTag, `langres`, `Name,Lang`},
		`MenuGroup`:  {defaultTag, defaultTag, `menugroup`, `Title,Body,Icon`},
		`MenuItem`:   {defaultTag, defaultTag, `menuitem`, `Title,Page,PageParams,Icon`},
		`SetVar`:     {setvarTag, defaultTag, `setvar`, `Name,Value`},
		`Strong`:     {defaultTag, defaultTag, `strong`, `Body,Class`},
	}
	tails = map[string]forTails{
		`button`: {map[string]tailInfo{
			`Alert`: {tplFunc{alertTag, defaultTailFull, `alert`, `Text,ConfirmButton,CancelButton,Icon`}, true},
			`Style`: {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
		}},
		`div`: {map[string]tailInfo{
			`Style`: {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
		}},
		`form`: {map[string]tailInfo{
			`Style`: {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
		}},
		`if`: {map[string]tailInfo{
			`Else`: {tplFunc{elseTag, elseFull, `else`, `Body`}, true},
		}},
		`image`: {map[string]tailInfo{
			`Style`: {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
		}},
		`input`: {map[string]tailInfo{
			`Validate`: {tplFunc{validateTag, validateFull, `validate`, `*`}, false},
			`Style`:    {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
		}},
		`label`: {map[string]tailInfo{
			`Style`: {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
		}},
		`linkpage`: {map[string]tailInfo{
			`Style`: {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
		}},
		`data`: {map[string]tailInfo{
			`Custom`: {tplFunc{customTag, defaultTailFull, `custom`, `Column,Body`}, false},
		}},
		`dbfind`: {map[string]tailInfo{
			`Columns`:   {tplFunc{tailTag, defaultTailFull, `columns`, `Columns`}, false},
			`Where`:     {tplFunc{tailTag, defaultTailFull, `where`, `Where`}, false},
			`WhereId`:   {tplFunc{tailTag, defaultTailFull, `whereid`, `WhereId`}, false},
			`Order`:     {tplFunc{tailTag, defaultTailFull, `order`, `Order`}, false},
			`Limit`:     {tplFunc{tailTag, defaultTailFull, `limit`, `Limit`}, false},
			`Offset`:    {tplFunc{tailTag, defaultTailFull, `offset`, `Offset`}, false},
			`Ecosystem`: {tplFunc{tailTag, defaultTailFull, `ecosystem`, `Ecosystem`}, false},
			`Custom`:    {tplFunc{customTag, defaultTailFull, `custom`, `Column,Body`}, false},
		}},
		`p`: {map[string]tailInfo{
			`Style`: {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
		}},
		`span`: {map[string]tailInfo{
			`Style`: {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
		}},
		`table`: {map[string]tailInfo{
			`Style`: {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
		}},
		`select`: {map[string]tailInfo{
			`Validate`: {tplFunc{validateTag, validateFull, `validate`, `*`}, false},
			`Style`:    {tplFunc{tailTag, defaultTailFull, `style`, `Style`}, false},
		}},
	}
	modes = [][]rune{{'(', ')'}, {'{', '}'}}
)

func init() {
	funcs[`Button`] = tplFunc{buttonTag, buttonTag, `button`, `Body,Page,Class,Contract,Params,PageParams`}
	funcs[`Div`] = tplFunc{defaultTailTag, defaultTailTag, `div`, `Class,Body`}
	funcs[`Form`] = tplFunc{defaultTailTag, defaultTailTag, `form`, `Class,Body`}
	funcs[`If`] = tplFunc{ifTag, ifFull, `if`, `Condition,Body`}
	funcs[`Image`] = tplFunc{defaultTailTag, defaultTailTag, `image`, `Src,Alt,Class`}
	funcs[`Include`] = tplFunc{includeTag, defaultTag, `include`, `Name`}
	funcs[`Input`] = tplFunc{defaultTailTag, defaultTailTag, `input`, `Name,Class,Placeholder,Type,Value`}
	funcs[`Label`] = tplFunc{defaultTailTag, defaultTailTag, `label`, `Body,Class,For`}
	funcs[`LinkPage`] = tplFunc{defaultTailTag, defaultTailTag, `linkpage`, `Body,Page,Class,PageParams`}
	funcs[`Data`] = tplFunc{dataTag, defaultTailTag, `data`, `Source,Columns,Data`}
	funcs[`DBFind`] = tplFunc{dbfindTag, defaultTailTag, `dbfind`, `Name,Source`}
	funcs[`And`] = tplFunc{andTag, defaultTag, `and`, `*`}
	funcs[`Or`] = tplFunc{orTag, defaultTag, `or`, `*`}
	funcs[`P`] = tplFunc{defaultTailTag, defaultTailTag, `p`, `Body,Class`}
	funcs[`Span`] = tplFunc{defaultTailTag, defaultTailTag, `span`, `Body,Class`}
	funcs[`Table`] = tplFunc{tableTag, defaultTailTag, `table`, `Source,Columns`}
	funcs[`Select`] = tplFunc{defaultTailTag, defaultTailTag, `select`, `Name,Source,Column,Class`}

	tails[`if`].Tails[`ElseIf`] = tailInfo{tplFunc{elseifTag, elseifFull, `elseif`, `Condition,Body`}, false}

}

func defaultTag(par parFunc) string {
	setAllAttr(par)
	par.Owner.Children = append(par.Owner.Children, par.Node)
	return ``
}

func addressTag(par parFunc) string {
	idval := (*par.Pars)[`Wallet`]
	if len(idval) == 0 {
		idval = (*par.Vars)[`wallet`]
	}
	id, _ := strconv.ParseInt(idval, 10, 64)
	if id == 0 {
		return `unknown address`
	}
	return converter.AddressToString(id)
}

func langresTag(par parFunc) string {
	lang := (*par.Pars)[`Lang`]
	if len(lang) == 0 {
		lang = (*par.Vars)[`accept_lang`]
	}
	ret, _ := language.LangText((*par.Pars)[`Name`], int(converter.StrToInt64((*par.Vars)[`state`])), lang)
	return ret
}

func andTag(par parFunc) string {
	count := len(*par.Pars)
	for i := 0; i < count; i++ {
		if !ifValue((*par.Pars)[strconv.Itoa(i)], par.Vars) {
			return `0`
		}
	}
	return `1`
}

func orTag(par parFunc) string {
	count := len(*par.Pars)
	for i := 0; i < count; i++ {
		if ifValue((*par.Pars)[strconv.Itoa(i)], par.Vars) {
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
				out, err := json.Marshal(par.Node.Attr[`custombody`].([][]*node)[i-defcol])
				if err == nil {
					ival = replace(string(out), 0, &vals)
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
	par.Node.Attr[`data`] = &data
	par.Owner.Children = append(par.Owner.Children, par.Node)
	return ``
}

func dbfindTag(par parFunc) string {
	var (
		fields string
		state  int64
	)
	if len((*par.Pars)[`Name`]) == 0 {
		return ``
	}
	defaultTail(par, `dbfind`)
	where := ``
	order := ``
	limit := 25
	if par.Node.Attr[`columns`] != nil {
		fields = converter.Escape(par.Node.Attr[`columns`].(string))
	}
	if len(fields) == 0 {
		fields = `*`
	}
	if par.Node.Attr[`where`] != nil {
		where = ` where ` + converter.Escape(par.Node.Attr[`where`].(string))
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
	if par.Node.Attr[`ecosystem`] != nil {
		state = converter.StrToInt64(par.Node.Attr[`ecosystem`].(string))
	} else {
		state = converter.StrToInt64((*par.Vars)[`state`])
	}
	tblname := fmt.Sprintf(`"%d_%s"`, state, strings.Trim(converter.EscapeName((*par.Pars)[`Name`]), `"`))
	list, err := model.GetAll(`select `+fields+` from `+tblname+where+order, limit)
	if err != nil {
		return err.Error()
	}
	/*	list := []map[string]string{{"id": "1", "amount": "200"}, {"id": "2", "amount": "300"}}
		fmt.Println(tblname, where, order)*/
	data := make([][]string, 0)
	cols := make([]string, 0)
	lencol := 0
	defcol := 0
	for _, item := range list {
		if lencol == 0 {
			for key := range item {
				cols = append(cols, key)
			}
			defcol = len(cols)
			if par.Node.Attr[`customs`] != nil {
				for _, v := range par.Node.Attr[`customs`].([]string) {
					cols = append(cols, v)
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
			} else {
				out, err := json.Marshal(par.Node.Attr[`custombody`].([][]*node)[i-defcol])
				if err == nil {
					ival = replace(string(out), 0, &item)
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
	par.Node.Attr[`data`] = &data
	par.Owner.Children = append(par.Owner.Children, par.Node)
	return ``
}

func customTag(par parFunc) string {
	setAllAttr(par)
	if par.Owner.Attr[`customs`] == nil {
		par.Owner.Attr[`customs`] = make([]string, 0)
		par.Owner.Attr[`custombody`] = make([][]*node, 0)
	}
	par.Owner.Attr[`customs`] = append(par.Owner.Attr[`customs`].([]string), par.Node.Attr[`column`].(string))
	par.Owner.Attr[`custombody`] = append(par.Owner.Attr[`custombody`].([][]*node), par.Node.Children)

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
	if len((*par.Pars)[`Name`]) >= 0 && len((*par.Vars)[`_include`]) < 5 {
		pattern, err := model.Single(`select value from "`+(*par.Vars)[`state`]+`_blocks" where name=?`, (*par.Pars)[`Name`]).String()
		if err != nil {
			return err.Error()
		}
		if len(pattern) > 0 {
			root := node{}
			(*par.Vars)[`_include`] += `1`
			process(pattern, &root, par.Vars)
			(*par.Vars)[`_include`] = (*par.Vars)[`_include`][:len((*par.Vars)[`_include`])-1]
			for _, item := range root.Children {
				par.Owner.Children = append(par.Owner.Children, item)
			}
		}
	}
	return ``
}

func setvarTag(par parFunc) string {
	if len((*par.Pars)[`Name`]) > 0 {
		(*par.Vars)[(*par.Pars)[`Name`]] = (*par.Pars)[`Value`]
	}
	return ``
}

func getvarTag(par parFunc) string {
	if len((*par.Pars)[`Name`]) > 0 {
		return macro((*par.Vars)[(*par.Pars)[`Name`]], par.Vars)
	}
	return ``
}

func tableTag(par parFunc) string {
	defaultTag(par)
	defaultTail(par, `table`)
	if len((*par.Pars)[`Columns`]) > 0 {
		imap := make(map[string]string)
		for _, v := range strings.Split((*par.Pars)[`Columns`], `,`) {
			v = strings.TrimSpace(v)
			if off := strings.IndexByte(v, '='); off == -1 {
				imap[v] = v
			} else {
				imap[strings.TrimSpace(v[:off])] = strings.TrimSpace(v[off+1:])
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
			curFunc := tails[tag].Tails[name].tplFunc
			pars := (*v)[:len(*v)-1]
			callFunc(&curFunc, par.Node, par.Vars, &pars, nil)
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
	if len((*par.Pars)[`Params`]) > 0 {
		imap := make(map[string]string)
		for _, v := range strings.Split((*par.Pars)[`Params`], `,`) {
			v = strings.TrimSpace(v)
			if off := strings.IndexByte(v, '='); off == -1 {
				imap[v] = v
			} else {
				imap[strings.TrimSpace(v[:off])] = strings.TrimSpace(v[off+1:])
			}
		}
		if len(imap) > 0 {
			par.Node.Attr[`params`] = imap
		}
	}
	defaultTail(par, `button`)
	return ``
}

func ifTag(par parFunc) string {
	cond := ifValue((*par.Pars)[`Condition`], par.Vars)
	if cond {
		for _, item := range par.Node.Children {
			par.Owner.Children = append(par.Owner.Children, item)
		}
	}
	if !cond && par.Tails != nil {
		for _, v := range *par.Tails {
			name := (*v)[len(*v)-1]
			curFunc := tails[`if`].Tails[name].tplFunc
			pars := (*v)[:len(*v)-1]
			callFunc(&curFunc, par.Owner, par.Vars, &pars, nil)
			if (*par.Vars)[`_cond`] == `1` {
				(*par.Vars)[`_cond`] = `0`
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
			curFunc := tails[`if`].Tails[name].tplFunc
			pars := (*v)[:len(*v)-1]
			callFunc(&curFunc, par.Node, par.Vars, &pars, nil)
		}
	}
	return ``
}

func elseifTag(par parFunc) string {
	cond := ifValue((*par.Pars)[`Condition`], par.Vars)
	if cond {
		for _, item := range par.Node.Children {
			par.Owner.Children = append(par.Owner.Children, item)
		}
		(*par.Vars)[`_cond`] = `1`
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
