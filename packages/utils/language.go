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
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

type cacheLang struct {
	res map[string]*map[string]string
}

var (
	LangList []string
	lang     = make(map[int]*cacheLang)
)

func IsLang(code string) bool {
	if LangList == nil {
		return true
	}
	for _, val := range LangList {
		if val == code {
			return true
		}
	}
	return false
}

func DefLang() string {
	if LangList == nil {
		return `en`
	}
	return LangList[0]
}

func UpdateLang(state int, name, value string) {
	if _, ok := lang[state]; !ok {
		return
	}
	var ires map[string]string
	json.Unmarshal([]byte(value), &ires)
	if len(ires) > 0 {
		(*lang[state]).res[name] = &ires
	}
}

func loadLang(state int) error {
	list, err := DB.GetAll(fmt.Sprintf(`select * from "%d_languages"`, state), -1)
	if err != nil {
		return err
	}
	res := &cacheLang{make(map[string]*map[string]string)}
	for _, ilist := range list {
		var ires map[string]string
		json.Unmarshal([]byte(ilist[`res`]), &ires)
		(*res).res[ilist[`name`]] = &ires
	}
	lang[state] = res
	//	fmt.Println(`Res`, *res)
	return nil
}

func LangText(in string, state int, accept string) (string,bool) {
	if strings.IndexByte(in, ' ') >= 0 {
		return in, false
	}
	if _, ok := lang[state]; !ok {
		if err := loadLang(state); err != nil {
			return err.Error(), false
		}
	}
	if lres, ok := (*lang[state]).res[in]; ok {
		langs := strings.Split(accept, `,`)
		lng := DefLang()
		for _, val := range langs {
			if len(val) < 2 {
				break
			}
			if !IsLang(val[:2]) {
				continue
			}
			if _, ok := (*lres)[val[:2]]; ok {
				lng = val[:2]
				break
			}
		}
		return (*lres)[lng], true
	}
	return in, false
}

func LangMacro(input string, state int, accept string) string {
	if len(input) == 0 {
		return input
	}
	syschar := '$'
	length := utf8.RuneCountInString(input)
	result := make([]rune, 0, length)
	isName := false
	name := make([]rune, 0, 128)
	clearname := func() {
		result = append(append(result, syschar), name...)
		isName = false
		name = name[:0]
	}
	for _, r := range input {
		if r != syschar {
			if isName {
				name = append(name, r)
				if len(name) > 64 || r < ' ' {
					clearname()
				}
			} else {
				result = append(result, r)
			}
			continue
		}
		if isName {
			value, ok := LangText(string(name), state, accept)
			if ok {
				result = append(result, []rune(value)...)
				isName = false
			} else {
				result = append(append(result, syschar), name...)
			}
			name = name[:0]
		} else {
			isName = true
		}
	}
	if isName {
		result = append(append(result, syschar), name...)
	}

	return string(result)
}
