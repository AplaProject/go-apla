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

package textproc

import (
	//"fmt"
	"strings"
	"unicode/utf8"
)

type TextFunc func(*map[string]string, ...string) string
type MapFunc func(*map[string]string, *map[string]string) string

type TextProc struct {
	syschar rune
	funcs   map[string]TextFunc
	maps    map[string]MapFunc
}

var (
	engine TextProc
)

func init() {
	engine = TextProc{syschar: '#', maps: make(map[string]MapFunc)}
	engine.funcs = map[string]TextFunc{
		`BR`:   Break,
		`Link`: Link,
		`Tag`:  Tag,
	}
}

func replace(input string, level int, vars *map[string]string) string {
	if len(input) == 0 {
		return input
	}
	length := utf8.RuneCountInString(input)
	result := make([]rune, 0, length)
	isName := false
	name := make([]rune, 0, 128)
	for _, r := range input {
		if r != engine.syschar {
			if isName {
				name = append(name, r)
				if len(name) > 64 || r < ' ' {
					result = append(append(result, engine.syschar), name...)
					isName = false
					name = name[:0]
				}
			} else {
				result = append(result, r)
			}
			continue
		}
		if isName {
			if value, ok := (*vars)[string(name)]; ok {
				if level < 10 {
					value = replace(value, level+1, vars)
				}
				result = append(result, []rune(value)...)
				isName = false
			} else {
				result = append(append(result, engine.syschar), name...)
			}
			name = name[:0]
		} else {
			isName = true
		}
	}
	if isName {
		result = append(append(result, engine.syschar), name...)
	}
	return string(result)
}

func AddMaps(funcs *map[string]MapFunc) {
	for key, ifunc := range *funcs {
		engine.maps[key] = ifunc
	}
}

func AddFuncs(funcs *map[string]TextFunc) {
	for key, ifunc := range *funcs {
		engine.funcs[key] = ifunc
	}
}

func Macro(input string, vars *map[string]string) string {
	return replace(input, 0, vars)
}

func Split(input string) *[][]string {
	var isArray, Par int

	ret := make([][]string, 0)
	value := make([]rune, 0)
	list := make([]string, 0)
	for _, ch := range input {
		if ch == '[' {
			isArray++
			continue
		}
		if isArray == 2 {
			if (Par == 0 && ch == ',') || ch == ']' {
				list = append(list, string(value))
				value = value[:0]
			} else {
				if ch == '(' {
					Par++
				} else if ch == ')' {
					Par--
				}
				value = append(value, ch)
			}
		}
		if ch == ']' {
			if isArray == 2 {
				ret = append(ret, list)
				list = make([]string, 0) //list[:0]
			}
			isArray--
			continue
		}
	}
	return &ret
}

func funcProcess(name string, params [][]rune, vars *map[string]string) string {
	pars := make([]string, 0)
	for _, item := range params {
		ipar := strings.TrimSpace(string(item))
		val := Process(ipar, vars)
		if len(val) == 0 {
			val = Macro(ipar, vars)
		}
		pars = append(pars, val)
	}
	return engine.funcs[name](vars, pars...)
}

func mapProcess(name string, params *map[string]string, vars *map[string]string) string {
	pars := make(map[string]string, 0)
	for key, item := range *params {
		var val string
		//		ipar := strings.TrimSpace(string(item))
		if len(item) > 0 && item[0] != '[' {
			val = Process(item, vars)
			if len(val) == 0 {
				val = Macro(item, vars)
			}
		} else {
			val = string(item)
		}
		pars[key] = val
	}
	return engine.maps[name](vars, &pars)
}

func Process(input string, vars *map[string]string) (out string) {
	var (
		isFunc, isMap, isArr int
		params               [][]rune
		pmap                 map[string]string
		isKey, toLine        bool
		pair                 rune
	)

	name := make([]rune, 0, 128)
	key := make([]rune, 0, 128)
	value := make([]rune, 0, 128)
	for _, ch := range input {
		if isMap > 0 {
			if pair > 0 {
				if ch != pair {
					value = append(value, ch)
				} else {
					pair = 0
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
				//				if isFunc == 0 {
				pmap[strings.TrimSpace(string(key))] = strings.TrimSpace(string(value))
				out += mapProcess(string(name), &pmap, vars) //+ "\r\n"
				name = name[:0]
				//				}
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
					pair = 0
				}
				continue
			}
			if len(params[len(params)-1]) == 0 && ch != ')' {
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
					out += funcProcess(string(name), params, vars) //+ "\r\n"
					name = name[:0]
					isFunc = 0
				} else {
					params[0] = append(params[0], ch)
				}
			} else {
				if ch == ')' {
					isFunc--
					if isFunc == 0 {
						out += funcProcess(string(name), params, vars) //+ "\r\n"
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
		if ch == 0xa && len(out) > 0 {
			out += "\n"
		}
		/*		if ch == 0xd && len(out) > 0 {
				out += "\r"
			}*/
		if ch < '!' {
			continue
		}
		if ch == '(' || ch == ':' {
			if _, ok := engine.funcs[string(name)]; !ok {
				return
			}
			params = make([][]rune, 1)
			params[0] = make([]rune, 0)
			isFunc++
			toLine = ch == ':'
		} else if ch == '{' {
			if _, ok := engine.maps[string(name)]; !ok {
				return
			}
			pmap = make(map[string]string)
			isKey = true
			key = key[:0]
			isMap++
		} else {
			name = append(name, ch)
			if len(name) > 64 {
				return
			}
		}
	}
	if toLine && isFunc > 0 {
		out += funcProcess(string(name), params, vars) //+ "\r\n"
	}
	return
}
