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
	//	"fmt"
	"unicode/utf8"
)

type TextFunc func(args ...string) string
type JsonFunc func(json string) string

type TextProc struct {
	syschar rune
	funcs   map[string]TextFunc
	jsons   map[string]JsonFunc
}

var (
	engine TextProc
)

func init() {
	engine = TextProc{'#', make(map[string]TextFunc), make(map[string]JsonFunc)}
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
				if len(name) > 64 {
					result = append(append(result, engine.syschar), name...)
					isName = false
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

func Do(input string, vars *map[string]string) string {
	return replace(input, 0, vars)
}
