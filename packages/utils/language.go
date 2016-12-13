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
)

type cacheLang struct {
	res map[string]*map[string]string
}

var (
	lang = make(map[int]*cacheLang)
)

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

func LangText(in string, state int, accept string) string {
	if strings.IndexByte(in, ' ') >= 0 {
		return in
	}
	if _, ok := lang[state]; !ok {
		if err := loadLang(state); err != nil {
			return err.Error()
		}
	}
	if lres, ok := (*lang[state]).res[in]; ok {
		langs := strings.Split(accept, `,`)
		lng := `en`
		for _, val := range langs {
			if _, ok := (*lres)[val[:2]]; ok {
				lng = val[:2]
				break
			}
		}
		return (*lres)[lng]
	}
	return in
}
