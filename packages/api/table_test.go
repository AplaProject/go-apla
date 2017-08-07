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

package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"
)

func TestTable(t *testing.T) {

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	name := randName(`table`)
	for _, glob := range []global{{`?limit=-1`, `0`}, {`?limit=-1&global=1`, `1`}} {
		value, err := json.Marshal([][]string{{`mytext`, `text`, `0`}, {`mynum`, `int64`, `1`}})
		if err != nil {
			t.Error(err)
			return
		}
		form := url.Values{"name": {name}, "columns": {string(value)},
			`global`: {glob.value}}
		if err := postTx(`table`, &form); err != nil {
			t.Error(err)
			return
		}
		ret, err := sendGet(`tables`+glob.url, nil)
		if err != nil {
			t.Error(err)
			return
		}
		var isTable bool
		for _, item := range ret[`list`].([]interface{}) {
			mitem := item.(map[string]interface{})
			get := mitem[`name`].(string)
			if get[strings.IndexByte(get, '_')+1:] == name {
				isTable = true
				break
			}
		}
		if !isTable {
			t.Error(fmt.Errorf(`unknown table %s`, name))
			return
		}
		prefix := `1`
		if glob.value == `1` {
			prefix = `global`
		}
		form = url.Values{"insert": {`true`}, "new_column": {`true`}, `general_update`: {`ContractConditions("MainCondition")`}}
		if err := putTx(`table/`+prefix+`_`+name, &form); err != nil {
			t.Error(err)
			return
		}
		form = url.Values{"permissions": {`false`}}
		if err := putTx(`column/`+prefix+`_`+name+`/mynum`, &form); err != nil {
			t.Error(err)
			return
		}
		form = url.Values{"name": {`newcol`}, "type": {`int64`},
			`index`: {`1`}, "permissions": {`true`}}
		if err := postTx(`column/`+prefix+`_`+name, &form); err != nil {
			t.Error(err)
			return
		}

		ret, err = sendGet(`table/`+name+glob.url, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if ret[`general_update`].(string) != `ContractConditions("MainCondition")` {
			t.Error(fmt.Errorf(`general_update is not right %s`, ret[`general_update`].(string)))
			return
		}
		if ret[`insert`].(string) != `true` {
			t.Error(fmt.Errorf(`insert is not right %s`, ret[`insert`].(string)))
			return
		}
		var isCol, isNewCol bool
		for _, item := range ret[`columns`].([]interface{}) {
			mitem := item.(map[string]interface{})
			if mitem[`name`].(string) == `mynum` {
				isCol = true
				if mitem[`perm`].(string) != `false` {
					t.Error(fmt.Errorf(`column permission is not right %s`, mitem[`perm`].(string)))
					return
				}
			}
			if mitem[`name`].(string) == `newcol` {
				isNewCol = true
				if mitem[`type`].(string) != `numbers` {
					t.Error(fmt.Errorf(`column type is not right %s`, mitem[`type`].(string)))
					return
				}
			}
		}
		if !isNewCol {
			t.Error(fmt.Errorf(`unknown new column`))
			return
		}
		if !isCol {
			t.Error(fmt.Errorf(`unknown column`))
			return
		}
	}
}
