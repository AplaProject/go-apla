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
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

type templ struct {
	Type     string                 `json:"type"`
	Name     string                 `json:"name"`
	Data     []interface{}          `json:"data,omitempty"`
	MapData  map[string]interface{} `json:"map,omitempty"`
	Children []*templ               `json:"children,omitempty"`
}

type templStack struct {
	End  string
	Prev *[]*templ
}

var (
	mfuncs  map[string]bool
	mblocks map[string]string
	mmaps   map[string]bool

	maps   = []string{`Table`, `TxForm`, `TxButton`, `ChartPie`, `ChartBar`}
	blocks = []string{
		`Divs,DivsEnd`, `UList,UListEnd`, `LiBegin,LiEnd`, `If,IfEnd`, `ForList,ForListEnd`,
		`Form,FormEnd`, `MenuGroup,MenuEnd`, `AutoUpdate,AutoUpdateEnd`,
	}
	funcs = []string{`Address`, `BtnEdit`, `InputMap`, `InputMapPoly`,
		`Image`, `ImageInput`, `Div`, `P`, `Em`, `Small`, `A`, `Span`, `Strong`,
		`LiTemplate`, `LinkPage`, `BtnPage`, `Li`,
		`CmpTime`, `Title`, `MarkDown`, `Navigation`, `PageTitle`,
		`PageEnd`, `StateVal`, `Json`, `And`, `Or`,
		`TxId`, `SetVar`, `GetList`, `GetRow`, `GetOne`, `TextHidden`,
		`ValueById`, `FullScreen`, `Ring`, `WiBalance`, `GetVar`,
		`WiAccount`, `WiCitizen`, `Map`, `MapPoint`, `StateLink`, `IfParams`,
		`Else`, `ElseIf`, `Trim`, `Date`, `DateTime`, `Now`, `Input`,
		`Textarea`, `InputMoney`, `InputAddress`,
		`BlockInfo`, `Back`, `ListVal`, `Tag`, `BtnContract`,
		`Label`, `Legend`, `Select`, `Param`, `Mult`,
		`Money`, `Source`, `Val`, `Lang`, `LangJS`, `InputDate`,
		`MenuItem`, `MenuPage`, `MenuBack`,
		`WhiteMobileBg`, `Bin2Hex`, `MessageBoard`, `Include`,
	}
)

func init() {
	mfuncs = make(map[string]bool)
	mmaps = make(map[string]bool)
	mblocks = make(map[string]string)
	for _, item := range funcs {
		mfuncs[item] = true
	}
	for _, item := range blocks {
		pars := strings.Split(item, `,`)
		if len(pars) == 2 {
			mblocks[pars[0]] = pars[1]
		}
	}
	for _, item := range maps {
		mmaps[item] = true
	}
}

func replace(input string, owner *[]interface{}) {
	if len(input) == 0 {
		return
	}
	length := utf8.RuneCountInString(input)
	result := make([]rune, 0, length)
	isName := false
	syschar := '#'
	isFunc := 0
	isMap := 0
	name := make([]rune, 0, 128)
	clearname := func() {
		result = append(append(result, syschar), name...)
		isName = false
		name = name[:0]
	}
	for _, r := range input {
		if r != syschar || isFunc > 0 {
			if isName {
				name = append(name, r)
				if r == '(' && isMap == 0 {
					if isFunc == 0 {
						if _, ok := mfuncs[string(name[:len(name)-1])]; !ok {
							clearname()
							continue
						}
					}
					isFunc++
				} else if r == '{' && isFunc == 0 {
					if isMap == 0 {
						if _, ok := mmaps[string(name[:len(name)-1])]; !ok {
							clearname()
							continue
						}
					}
					isMap++
				} else if r == ')' && isFunc > 0 {
					if isFunc--; isFunc == 0 {
						if len(result) > 0 {
							*owner = append(*owner, string(result))
							result = result[:0]
						}
						powner := make([]*templ, 0)
						processTemplate(string(name), &powner)
						for _, item := range powner {
							*owner = append(*owner, item)
						}
						isName = false
						name = name[:0]
					}
				} else if r == '}' && isMap > 0 {
					if isMap--; isMap == 0 {
						if len(result) > 0 {
							*owner = append(*owner, string(result))
							result = result[:0]
						}
						powner := make([]*templ, 0)
						processTemplate(string(name), &powner)
						for _, item := range powner {
							*owner = append(*owner, item)
						}
						isName = false
						name = name[:0]
					}
				} else if (len(name) > 64 && isFunc == 0) || r < ' ' || (r == ' ' && isFunc == 0 && isMap == 0) {
					clearname()
				}
			} else {
				if r == '(' {
					name = name[:0]
					for i := len(result) - 1; i >= 0; i-- {
						if (result[i] >= 'a' && result[i] <= 'z') ||
							(result[i] >= 'A' && result[i] <= 'Z') {
							name = append(name, result[i])
						} else {
							break
						}
					}
					if len(name) > 0 {
						for i, j := 0, len(name)-1; i < j; i, j = i+1, j-1 {
							name[i], name[j] = name[j], name[i]
						}
						if _, ok := mfuncs[string(name)]; ok {
							isName = true
							isFunc++
							result = result[:len(result)-len(name)]
							name = append(name, '(')
							continue
						}
						name = name[:0]
					}

				}
				result = append(result, r)
			}
			continue
		}
		if isName {
			/*			if value, ok := (*vars)[string(name)]; ok {
						if level < 10 {
							value = replace(value, level+1, vars)
						}
						result = append(result, []rune(value)...)
						isName = false
					} else {*/
			result = append(append(result, syschar), name...)
			//			}
			name = name[:0]
		} else {
			isName = true
		}
	}
	if isName {
		result = append(append(result, syschar), name...)
	}
	if len(result) > 0 {
		*owner = append(*owner, string(result))
		result = result[:0]
	}
}

func funcProcess(name string, params [][]rune, owner *[]*templ) bool {
	pars := make([]interface{}, 0)
	for _, item := range params {
		ipar := strings.TrimSpace(string(item))
		off := strings.Index(ipar, `#=`)
		if off < 0 || off != strings.Index(ipar, `#`) {
			tmp := make([]*templ, 0)
			if strings.ContainsAny(ipar, `({:`) && processTemplate(ipar, &tmp) == nil {
				pars = append(pars, tmp)
			} else {
				itmp := make([]interface{}, 0)
				replace(ipar, &itmp)
				if len(itmp) > 1 {
					pars = append(pars, itmp)
				} else if len(itmp) == 1 {
					pars = append(pars, ipar)
				}
			}
		} else {
			pars = append(pars, ipar)
		}
	}
	itype := `fn`
	if _, ok := mblocks[name]; ok {
		itype = `block`
	}
	*owner = append(*owner, &templ{Type: itype, Name: name, Data: pars})
	return itype == `block`
}

func mapProcess(name string, params *map[string]string, owner *[]*templ) {
	pars := make(map[string]interface{})
	for key, item := range *params {
		var val string
		/*		if len(item) > 0 && item[0] != '[' {
				val = item
			} else {*/
		val = item
		pars[key] = val
	}
	*owner = append(*owner, &templ{Type: `map`, Name: name, MapData: pars})
}

func processTemplate(input string, owner *[]*templ) error {
	var (
		isFunc, isMap, isArr int
		params               [][]rune
		pmap                 map[string]string
		isKey, toLine, skip  bool
		pair                 rune
	)

	stackBlocks := make([]templStack, 0)
	noproc := true
	name := make([]rune, 0, 128)
	key := make([]rune, 0, 128)
	value := make([]rune, 0, 128)
	for off, ch := range input {
		if skip {
			skip = false
			continue
		}
		if isMap > 0 {
			if pair > 0 {
				if ch != pair {
					value = append(value, ch)
				} else {
					if off+1 == len(input) || rune(input[off+1]) != pair {
						pair = 0
					} else {
						value = append(value, ch)
						skip = true
					}
				}
				continue
			}
			if !isKey && len(value) == 0 {
				if ch >= '!' {
					if ch == '"' || ch == '`' {
						pair = ch
					} else {
						if ch == '[' {
							isArr++
						}
						value = append(value, ch)
					}
				}
				continue
			}
			if ch == '}' {
				isMap--
				pmap[strings.TrimSpace(string(key))] = strings.TrimSpace(string(value))
				mapProcess(string(name), &pmap, owner)
				name = name[:0]
			}
			if isKey {
				if ch < '!' {
					continue
				}
				if isKey && ch == ':' {
					isKey = false
					value = value[:0]
					continue
				}
				key = append(key, ch)
				continue
			}
			if isArr == 0 && (ch == 0xa || ch == ',') {
				pmap[strings.TrimSpace(string(key))] = strings.TrimSpace(string(value))
				isKey = true
				key = key[:0]
				value = value[:0]
			}
			if ch == '[' {
				isArr++
			}
			if ch == ']' {
				isArr--
			}
			value = append(value, ch)
			continue
		}
		if isFunc > 0 {
			if pair > 0 {
				if ch != pair {
					params[len(params)-1] = append(params[len(params)-1], ch)
				} else {
					if off+1 == len(input) || rune(input[off+1]) != pair {
						pair = 0
					} else {
						params[len(params)-1] = append(params[len(params)-1], ch)
						skip = true
					}
				}
				continue
			}
			if len(params[len(params)-1]) == 0 && ch != ')' && ch != ',' && !toLine {
				if ch >= '!' {
					if ch == '"' || ch == '`' {
						pair = ch
					} else {
						params[len(params)-1] = append(params[len(params)-1], ch)
					}
				}
				continue
			}
			if toLine {
				if ch == 0xa {
					if funcProcess(string(name), params, owner) {
						stackBlocks = append(stackBlocks, templStack{End: mblocks[string(name)],
							Prev: owner})
						tmpowner := make([]*templ, 0)
						owner = &tmpowner
					}
					name = name[:0]
					isFunc = 0
				} else {
					params[0] = append(params[0], ch)
				}
			} else {
				if ch == ')' {
					isFunc--
					if isFunc == 0 {
						if funcProcess(string(name), params, owner) {
							stackBlocks = append(stackBlocks, templStack{End: mblocks[string(name)],
								Prev: owner})
							tmpowner := make([]*templ, 0)
							owner = &tmpowner
						}
						name = name[:0]
					}
				}
				if ch == '(' {
					isFunc++
				}
				if ch == ',' && isFunc == 1 {
					params = append(params, make([]rune, 0))
				} else {
					params[len(params)-1] = append(params[len(params)-1], ch)
				}
			}
			continue
		}
		if ch < '!' {
			continue
		}
		if ch == '(' || ch == ':' {
			if slen := len(stackBlocks); slen > 0 && stackBlocks[slen-1].End == string(name) {
				*owner = append(*owner, &templ{Type: `fn`, Name: string(name)})
				name = name[:0]
				children := owner
				owner = stackBlocks[slen-1].Prev
				(*owner)[len(*owner)-1].Children = *children
				stackBlocks = stackBlocks[:slen-1]
				continue
			}
			if _, ok := mblocks[string(name)]; !mfuncs[string(name)] && !ok {
				return fmt.Errorf(`Unknown function %s`, string(name))
			}
			noproc = false
			params = make([][]rune, 1)
			params[0] = make([]rune, 0)
			isFunc++
			toLine = ch == ':'
		} else if ch == '{' {
			if _, ok := mmaps[string(name)]; !ok {
				return fmt.Errorf(`Unknown map function %s`, string(name))
			}
			pmap = make(map[string]string)
			isKey = true
			noproc = false
			key = key[:0]
			isMap++
		} else {
			name = append(name, ch)
			if len(name) > 64 {
				return fmt.Errorf(`Too long name %s`, string(name))
			}
		}
	}
	if toLine && isFunc > 0 {
		funcProcess(string(name), params, owner)
	}
	if noproc {
		return fmt.Errorf(`Text must contains functions`)
	}
	return nil
}

// Template2JSON converts templates to JSON data
func Template2JSON(input string) []byte {
	root := make([]*templ, 0)

	err := processTemplate(input, &root)
	if err != nil {
		return []byte(err.Error())
	}
	out, err := json.Marshal(root)
	if err != nil {
		return []byte(err.Error())
	}
	return out
}
