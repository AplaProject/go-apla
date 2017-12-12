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

package language

import (
	"encoding/json"
	"strings"
	"unicode/utf8"

	"strconv"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

type cacheLang struct {
	res map[string]*map[string]string
}

var (
	// LangList is the list of available languages. It stores two-bytes codes
	LangList []string
	lang     = make(map[int]*cacheLang)
)

// IsLang checks if there is a language with code name
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

// DefLang returns the default language
func DefLang() string {
	if LangList == nil {
		return `en`
	}
	return LangList[0]
}

// UpdateLang updates language sources for the specified state
func UpdateLang(state int, name, value string, vde bool) {
	if vde {
		state = -state
	}
	if _, ok := lang[state]; !ok {
		lang[state] = &cacheLang{make(map[string]*map[string]string)}
	}
	var ires map[string]string
	err := json.Unmarshal([]byte(value), &ires)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "value": value, "error": err}).Error("Unmarshalling json")
	}
	if len(ires) > 0 {
		(*lang[state]).res[name] = &ires
	}
}

// loadLang download the language sources from database for the state
func loadLang(state int, vde bool) error {
	language := &model.Language{}
	prefix := strconv.FormatInt(int64(state), 10)
	if vde {
		prefix += `_vde`
	}
	languages, err := language.GetAll(prefix)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Error querying all languages")
		return err
	}
	list := make([]map[string]string, 0)
	for _, l := range languages {
		list = append(list, l.ToMap())
	}
	res := &cacheLang{make(map[string]*map[string]string)}
	for _, ilist := range list {
		var ires map[string]string
		err := json.Unmarshal([]byte(ilist[`res`]), &ires)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "value": ilist["res"], "error": err}).Error("Unmarshalling json")
		}
		(*res).res[ilist[`name`]] = &ires
	}
	if vde {
		state = -state
	}
	lang[state] = res
	return nil
}

// LangText looks for the specified word through language sources and returns the meaning of the source
// if it is found. Search goes according to the languages specified in 'accept'
func LangText(in string, state int, accept string, vde bool) (string, bool) {
	if strings.IndexByte(in, ' ') >= 0 || state == 0 {
		return in, false
	}
	istate := state
	if vde {
		istate = -state
	}
	if _, ok := lang[istate]; !ok {
		if err := loadLang(state, vde); err != nil {
			return err.Error(), false
		}
	}
	if lres, ok := (*lang[istate]).res[in]; ok {
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
		if len((*lres)[lng]) == 0 {
			for _, val := range *lres {
				return val, true
			}
		}
		return (*lres)[lng], true
	}
	return in, false
}

// LangMacro replaces all inclusions of $resname$ in the incoming text with the corresponding language resources,
// if they exist
func LangMacro(input string, state int, accept string, vde bool) string {
	if !strings.ContainsRune(input, '$') {
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
			value, ok := LangText(string(name), state, accept, vde)
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

// GetLang returns the first language from accept-language
func GetLang(state int, accept string) (lng string) {
	lng = DefLang()
	for _, val := range strings.Split(accept, `,`) {
		if len(val) < 2 {
			continue
		}
		if !IsLang(val[:2]) {
			continue
		}
		lng = val[:2]
		break
	}
	return
}
