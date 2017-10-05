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
	"fmt"
	"strings"
)

var (
	funcs = map[string]tplFunc{
		`Div`:       {defaultTag, defaultTag, `div`, `Class,Body`},
		`Em`:        {defaultTag, defaultTag, `em`, `Body,Class`},
		`Form`:      {defaultTag, defaultTag, `form`, `Class,Body`},
		`GetVar`:    {getvarTag, defaultTag, `getvar`, `Name`},
		`InputErr`:  {defaultTag, defaultTag, `inputerr`, `*`},
		`Label`:     {defaultTag, defaultTag, `label`, `Body,Class,For`},
		`MenuGroup`: {defaultTag, defaultTag, `menugroup`, `Title,Body,Icon`},
		`MenuItem`:  {defaultTag, defaultTag, `menuitem`, `Title,Page,PageParams,Icon`},
		`P`:         {defaultTag, defaultTag, `p`, `Body,Class`},
		`SetVar`:    {setvarTag, defaultTag, `setvar`, `Name,Value`},
		`Span`:      {defaultTag, defaultTag, `span`, `Body,Class`},
		`Strong`:    {defaultTag, defaultTag, `strong`, `Body,Class`},
		`Style`:     {defaultTag, defaultTag, `style`, `Css`},
	}
	tails = map[string]forTails{
		`button`: {map[string]tailInfo{
			`Alert`: {tplFunc{alertTag, alertFull, `alert`, `Text,ConfirmButton,CancelButton,Icon`}, true},
		}},
		`if`: {map[string]tailInfo{
			`Else`: {tplFunc{elseTag, elseFull, `else`, `Body`}, true},
		}},
		`input`: {map[string]tailInfo{
			`Validate`: {tplFunc{validateTag, validateFull, `validate`, `*`}, true},
		}},
		`dbfind`: {map[string]tailInfo{
			`Columns`: {tplFunc{tailTag, defaultTag, `columns`, `Columns`}, false},
			`Where`:   {tplFunc{tailTag, defaultTag, `where`, `Where`}, false},
			`WhereId`: {tplFunc{tailTag, defaultTag, `whereid`, `WhereId`}, false},
			`Order`:   {tplFunc{tailTag, defaultTag, `order`, `Order`}, false},
			`Count`:   {tplFunc{tailTag, defaultTag, `count`, `Count`}, false},
			`Offset`:  {tplFunc{tailTag, defaultTag, `offset`, `Offset`}, false},
		}},
	}
	modes = [][]rune{{'(', ')'}, {'{', '}'}}
)

func init() {
	funcs[`Button`] = tplFunc{buttonTag, buttonTag, `button`, `Body,Page,Class,Contract,Params,PageParams`}
	funcs[`If`] = tplFunc{ifTag, ifFull, `if`, `Condition,Body`}
	funcs[`Input`] = tplFunc{inputTag, inputTag, `input`, `Name,Class,Placeholder,Type,Value`}
	funcs[`DBFind`] = tplFunc{dbfindTag, defaultTag, `dbfind`, `Name`}

	tails[`if`].Tails[`ElseIf`] = tailInfo{tplFunc{elseifTag, elseifFull, `elseif`, `Condition,Body`}, false}

}

func defaultTag(par parFunc) string {
	setAllAttr(par)
	par.Owner.Children = append(par.Owner.Children, par.Node)
	return ``
}

func alertTag(par parFunc) string {
	setAllAttr(par)
	par.Owner.Attr[`alert`] = par.Node.Attr
	return ``
}

func alertFull(par parFunc) string {
	setAllAttr(par)
	par.Owner.Tail = append(par.Owner.Tail, par.Node)
	return ``
}

func dbfindTag(par parFunc) string {
	if len((*par.Pars)[`Name`]) == 0 {
		return ``
	}
	defaultTail(par, `dbfind`)
	fmt.Println(`Par`, par.Node.Attr)
	return ``
}

func tailTag(par parFunc) string {
	setAllAttr(par)
	for key, v := range par.Node.Attr {
		par.Owner.Attr[key] = v
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

func inputTag(par parFunc) string {
	defaultTag(par)
	defaultTail(par, `input`)
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
