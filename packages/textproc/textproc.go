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
	"strings"
	"unicode/utf8"
)

const (
	null = `NULL`
)

// TextFunc is a function type with 1 - variables, 2... - parameters
type TextFunc func(*map[string]string, ...string) string

// MapFunc is a function type with 1 - variables, 2 - map of parameters
type MapFunc func(*map[string]string, *map[string]string) string

// TextProc is a strcuture for the text processing
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
	isFunc := 0
	isMap := 0
	name := make([]rune, 0, 128)
	clearname := func() {
		result = append(append(result, engine.syschar), name...)
		isName = false
		name = name[:0]
	}
	for _, r := range input {
		if r != engine.syschar || isFunc > 0 {
			if isName {
				name = append(name, r)
				if r == '(' && isMap == 0 {
					if isFunc == 0 {
						if _, ok := engine.funcs[string(name[:len(name)-1])]; !ok {
							clearname()
							continue
						}
					}
					isFunc++
				} else if r == '{' && isFunc == 0 {
					if isMap == 0 {
						if _, ok := engine.maps[string(name[:len(name)-1])]; !ok {
							clearname()
							continue
						}
					}
					isMap++
				} else if r == ')' && isFunc > 0 {
					if isFunc--; isFunc == 0 {
						result = append(result, []rune(Process(string(name), vars))...)
						isName = false
						name = name[:0]
					}
				} else if r == '}' && isMap > 0 {
					if isMap--; isMap == 0 {
						result = append(result, []rune(Process(string(name), vars))...)
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
						if _, ok := engine.funcs[string(name)]; ok {
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

// AddMaps appends MapFunc functions to the text processing engine
func AddMaps(funcs *map[string]MapFunc) {
	for key, ifunc := range *funcs {
		engine.maps[key] = ifunc
	}
}

// AddFuncs appends TextFunc functions to the text processing engine
func AddFuncs(funcs *map[string]TextFunc) {
	for key, ifunc := range *funcs {
		engine.funcs[key] = ifunc
	}
}

// Macro replaces macro variables in the input string and returns the result
func Macro(input string, vars *map[string]string) string {
	return replace(input, 0, vars)
}

// Split parses the input string as [[,,],[,,],[,,]]
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
	if (strings.HasSuffix((*vars)[`ifs`], `0`) || strings.HasSuffix((*vars)[`ifs`], `-`)) && name != `If` && name != `IfEnd` && name != `Else` &&
		name != `ElseIf` {
		return ``
	}
	pars := make([]string, 0)
	for _, item := range params {
		var val string
		ipar := strings.TrimSpace(string(item))
		off := strings.Index(ipar, `#=`)
		if off < 0 || off != strings.Index(ipar, `#`) {
			val = Process(ipar, vars)
			if val == null {
				val = Macro(ipar, vars)
			}
		} else {
			val = ipar
		}
		pars = append(pars, val)
	}
	return engine.funcs[name](vars, pars...)
}

func mapProcess(name string, params *map[string]string, vars *map[string]string) string {
	if strings.HasSuffix((*vars)[`ifs`], `0`) || strings.HasSuffix((*vars)[`ifs`], `-`) {
		return ``
	}
	pars := make(map[string]string)
	for key, item := range *params {
		var val string
		//		ipar := strings.TrimSpace(string(item))
		if len(item) > 0 && item[0] != '[' {
			val = Process(item, vars)
			if val == null {
				val = Macro(item, vars)
			}
		} else {
			val = item
		}
		pars[key] = val
	}
	return engine.maps[name](vars, &pars)
}

// Process replaces variables and func calling in the input string and returns the result.
func Process(input string, vars *map[string]string) (out string) {
	var (
		isFunc, isMap, isArr int
		params               [][]rune
		pmap                 map[string]string
		isKey, toLine, skip  bool
		pair                 rune
	)
	noproc := true
	name := make([]rune, 0, 128)
	key := make([]rune, 0, 128)
	value := make([]rune, 0, 128)
	forbody := make([]rune, 0, 1024)
	autobody := make([]rune, 0, 1024)
	for off, ch := range input {
		if (*vars)[`auto_loop`] == `1` {
			if off+13 < len(input) && input[off:off+13] == `AutoUpdateEnd` {
				(*vars)[`auto_body`] = string(autobody)
				autobody = autobody[:0]
				(*vars)[`auto_loop`] = `0`
			} else {
				autobody = append(autobody, ch)
				//	continue
			}
		}
		if (*vars)[`for_loop`] == `1` {
			if off+10 < len(input) && input[off:off+10] == `ForListEnd` {
				(*vars)[`for_body`] = string(forbody)
				forbody = forbody[:0]
				(*vars)[`for_loop`] = `0`
			} else {
				forbody = append(forbody, ch)
				continue
			}
		}
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
						if string(name) == `Func` {
							name = []rune(funcProcess(string(name), params, vars))
						} else {
							out += funcProcess(string(name), params, vars) //+ "\r\n"
							name = name[:0]
						}
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
				return null
			}
			noproc = false
			params = make([][]rune, 1)
			params[0] = make([]rune, 0)
			isFunc++
			toLine = ch == ':'
		} else if ch == '{' {
			if _, ok := engine.maps[string(name)]; !ok {
				return null
			}
			pmap = make(map[string]string)
			isKey = true
			noproc = false
			key = key[:0]
			isMap++
		} else {
			name = append(name, ch)
			if len(name) > 64 {
				return null
			}
		}
	}
	if toLine && isFunc > 0 {
		out += funcProcess(string(name), params, vars) //+ "\r\n"
	}
	if noproc {
		return null
	}
	return
}
