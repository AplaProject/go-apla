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
	"encoding/json"
	"html"
	"strings"
	//	"unicode/utf8"
)

const (
	tagText = `text`
)

type node struct {
	Tag  string                 `json:"tag"`
	Attr map[string]interface{} `json:"attr,omitempty"`
	//	Map      map[string]map[string]string `json:"map,omitempty"`
	Text     string  `json:"text,omitempty"`
	Children []*node `json:"children,omitempty"`
	Tail     []*node `json:"tail,omitempty"`
}

type parFunc struct {
	Node *node
	Vars *map[string]string
	Pars *map[string]string
}

type nodeFunc func(par parFunc) string

type tplFunc struct {
	Func   nodeFunc // process function
	Tag    string   // HTML tag
	Params string   // names of parameters
}

type tailInfo struct {
	tplFunc
	Last bool
}

type forTails struct {
	Tails map[string]tailInfo
}

var (
	funcs = map[string]tplFunc{
		`Div`:    {defaultTag, `div`, `Class,Body`},
		`Button`: {buttonTag, `button`, `Body,Page,Class,Contract,Params,PageParams,Alert`},
		`Em`:     {defaultTag, `em`, `Body,Class`},
		`Form`:   {defaultTag, `form`, `Class,Body`},
		`If`:     {defaultTag, `if`, `Condition,Body`},
		`Input`:  {inputTag, `input`, `Name,Class,Placeholder,Type,Value,Validate`},
		`Label`:  {labelTag, `label`, `Body,Class,For`},
		`P`:      {defaultTag, `p`, `Body,Class`},
		`Span`:   {defaultTag, `span`, `Body,Class`},
		`Strong`: {defaultTag, `strong`, `Body,Class`},
	}
	tails = map[string]forTails{
		`if`: {map[string]tailInfo{
			`Else`:   {tplFunc{defaultTag, `else`, `Body`}, true},
			`ElseIf`: {tplFunc{defaultTag, `elseif`, `Condition,Body`}, false},
		}},
	}
	modes = [][]rune{{'(', ')'}, {'{', '}'}}
)

func setAttr(par parFunc, name string) {
	if len((*par.Pars)[name]) > 0 {
		par.Node.Attr[strings.ToLower(name)] = (*par.Pars)[name]
	}
}

func defaultTag(par parFunc) string {
	setAttr(par, `Class`)
	setAttr(par, `Name`)
	return ``
}

func buttonTag(par parFunc) string {
	defaultTag(par)
	setAttr(par, `Page`)
	setAttr(par, `Contract`)
	setAttr(par, `Alert`)
	setAttr(par, `PageParams`)
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
	return ``
}

func inputTag(par parFunc) string {
	defaultTag(par)
	setAttr(par, `Placeholder`)
	setAttr(par, `Value`)
	setAttr(par, `Validate`)
	setAttr(par, `Type`)
	return ``
}

func labelTag(par parFunc) string {
	defaultTag(par)
	setAttr(par, `For`)
	return ``
}

func appendText(owner *node, text string) {
	if len(strings.TrimSpace(text)) == 0 {
		return
	}
	if len(text) > 0 {
		owner.Children = append(owner.Children, &node{Tag: tagText, Text: html.EscapeString(text)})
	}
}

func callFunc(curFunc *tplFunc, owner *node, vars *map[string]string, params *[]string) {
	var curNode node
	pars := make(map[string]string)
	parFunc := parFunc{
		Vars: vars,
	}
	for i, v := range strings.Split(curFunc.Params, `,`) {
		if i < len(*params) {
			val := strings.TrimSpace((*params)[i])
			off := strings.IndexByte(val, ':')
			if off != -1 && strings.Contains(curFunc.Params, val[:off]) {
				pars[val[:off]] = strings.TrimSpace(val[off+1:])
			} else {
				pars[v] = val
			}
		} else if _, ok := pars[v]; !ok {
			pars[v] = ``
		}
	}
	if len(curFunc.Tag) > 0 {
		curNode.Tag = curFunc.Tag
		curNode.Attr = make(map[string]interface{})
		if len(pars[`Body`]) > 0 {
			process(pars[`Body`], &curNode, vars)
		}
		owner.Children = append(owner.Children, &curNode)
		parFunc.Node = &curNode
	}
	parFunc.Pars = &pars
	out := curFunc.Func(parFunc)
	if len(out) > 0 {
		if len(owner.Children) > 0 && owner.Children[len(owner.Children)-1].Tag == tagText {
			owner.Children[len(owner.Children)-1].Text += out
		} else {
			appendText(owner, out)
		}
	}
}

func process(input string, owner *node, vars *map[string]string) {
	var (
		nameOff int
		/*chOff,*/ pair rune
		params          []string
		curp            int
		skip, isFunc    bool
		curFunc         tplFunc
		level, mode     int
	)
	//	fmt.Println(`Input`, input)
	name := make([]rune, 0, 128)
main:
	for off, ch := range input {
		if skip {
			skip = false
			continue
		}
		if isFunc {
			//			fmt.Println(off, ch, curp, params)
			if pair > 0 {
				if ch != pair {
					params[curp] += string(ch)
				} else {
					if off+1 == len(input) || rune(input[off+1]) != pair {
						pair = 0
					} else {
						params[curp] += string(ch)
						skip = true
					}
				}
				continue
			}
			if len(params[curp]) == 0 && ch != modes[mode][1] && ch != ',' {
				if ch >= '!' {
					if ch == '"' || ch == '`' {
						pair = ch
					} else {
						params[curp] += string(ch)
					}
				}
				continue
			}
			//			fmt.Println(`CH`, string(ch))
			switch ch {
			case ',':
				if mode == 0 && level == 1 {
					params = append(params, ``)
					curp++
					continue
				}
			case modes[mode][0]:
				level++
			case modes[mode][1]:
				if level > 0 {
					level--
				}
				if level == 0 {
					if mode == 0 && off+1 < len(input) && rune(input[off+1]) == modes[1][0] &&
						strings.Contains(curFunc.Params, `Body`) {
						mode = 1
						params = append(params, `Body:`)
						curp++
						skip = true
						level = 1
						continue
					}
					if tail, ok := tails[curFunc.Tag]; ok && off+1 < len(input) && input[off+1] == '.' {
						for key, _ := range tail.Tails {
							if strings.HasPrefix(input[off+2:], key+`(`) || strings.HasPrefix(input[off+2:], key+`{`) {
								skip = true
								mode = 0
								level = 0
								continue main
							}
						}
					}
					callFunc(&curFunc, owner, vars, &params)
					isFunc = false
					continue
				}
			}
			params[curp] += string(ch)
			continue
		}
		if ch == '(' {
			if curFunc, isFunc = funcs[string(name[nameOff:])]; isFunc {
				params = make([]string, 1)
				curp = 0
				appendText(owner, string(name[:nameOff]))
				name = name[:0]
				nameOff = 0
				level = 1
				mode = 0
				continue
			}
		}
		if (ch < 'A' || ch > 'Z') && (ch < 'a' || ch > 'z') {
			nameOff = len(name) + 1
		}
		name = append(name, ch)
	}
	appendText(owner, string(name))
}

// Template2JSON converts templates to JSON data
func Template2JSON(input string) []byte {
	vars := make(map[string]string)
	root := node{}
	process(input, &root, &vars)
	out, err := json.Marshal(root.Children)
	if err != nil {
		return []byte(err.Error())
	}
	return out
}
