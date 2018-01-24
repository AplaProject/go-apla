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
	"fmt"
	"net/url"
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
	var (
		err     error
		ret     vdeCreateResult
		retCont contentResult
	)

	if err = keyLogin(1); err != nil {
		t.Error(err)
		return
	}

	if err = sendPost(`vde/create`, nil, &ret); err != nil &&
		err.Error() != `400 {"error": "E_VDECREATED", "msg": "Virtual Dedicated Ecosystem is already created" }` {
		t.Error(err)
		return
	}
	name := randName(`tbl`)
	form := url.Values{"vde": {`true`}, "Name": {name}, "Columns": {`[{"name":"my","type":"varchar", "index": "1", 
	  "conditions":"true"},
	{"name":"amount", "type":"number","index": "0", "conditions":"{\"update\":\"true\", \"read\":\"true\"}"},
	{"name":"active", "type":"character","index": "0", "conditions":"{\"update\":\"true\", \"read\":\"false\"}"}]`},
		"Permissions": {`{"insert": "true", "update" : "true", "read": "true", "new_column": "true"}`}}
	err = postTx(`NewTable`, &form)
	if err != nil {
		t.Error(err)
		return
	}

	contFill := fmt.Sprintf(`contract %s {
		action {
			DBInsert("%[1]s", "my,amount", "Alex", 100 )
			DBInsert("%[1]s", "my,amount", "Alex 2", 13300 )
			DBInsert("%[1]s", "my,amount", "Mike", 0 )
			DBInsert("%[1]s", "my,amount", "Mike 2", 25500 )
			DBInsert("%[1]s", "my,amount", "John Mike", 0 )
			DBInsert("%[1]s", "my,amount", "Serena Martin", 777 )
		}
	}

	contract Get%[1]s {
		action {
			var row array
			row = DBFind("%[1]s").Where("id>= ? and id<= ?", 2, 5)
		}
	}

	contract GetOK%[1]s {
		action {
			var row array
			row = DBFind("%[1]s").Columns("my,amount").Where("id>= ? and id<= ?", 2, 5)
		}
	}

	contract GetData%[1]s {
		action {
			var row array
			row = DBFind("%[1]s").Columns("active").Where("id>= ? and id<= ?", 2, 5)
		}
	}

	func ReadFilter%[1]s bool {
		var i int
		var row map
		while i < Len($data) {
			row = $data[i]
			if i == 1 || i == 3 {
				row["my"] = "No name"
				$data[i] = row
			}
			i = i+ 1
		}
		return true
	}
	`, name)
	form = url.Values{"Value": {contFill},
		"Conditions": {`true`}, "vde": {`true`}}
	if err := postTx(`NewContract`, &form); err != nil {
		t.Error(err)
		return
	}
	if err := postTx(name, &url.Values{"vde": {`true`}}); err != nil {
		t.Error(err)
		return
	}
	if err := postTx(`GetData`+name, &url.Values{"vde": {`true`}}); err.Error() != `500 {"error": "E_SERVER", "msg": "{\"type\":\"panic\",\"error\":\"Access denied\"}" }` {
		t.Errorf(`access problem`)
		return
	}
	err = sendPost(`content`, &url.Values{`vde`: {`true`}, `template`: {
		`DBFind(` + name + `, src).Limit(2)`}}, &retCont)
	if err != nil {
		t.Error(err)
		return
	}
	if strings.Contains(RawToString(retCont.Tree), `active`) {
		t.Errorf(`wrong tree %s`, RawToString(retCont.Tree))
		return
	}

	if err := postTx(`GetOK`+name, &url.Values{"vde": {`true`}}); err != nil {
		t.Error(err)
		return
	}
	if err := postTx(`EditColumn`, &url.Values{"vde": {`true`}, `TableName`: {name}, `Name`: {`active`},
		`Permissions`: {`{"update":"true", "read":"ContractConditions(\"MainCondition\")"}`}}); err != nil {
		t.Error(err)
		return
	}
	if err := postTx(`Get`+name, &url.Values{"vde": {`true`}}); err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"Name": {name}, "vde": {`true`},
		"Permissions": {`{"insert": "ContractConditions(\"MainCondition\")", 
		"update" : "true", "filter": "ReadFilter` + name + `()", "new_column": "ContractConditions(\"MainCondition\")"}`}}
	err = postTx(`EditTable`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	var tableInfo tableResult
	err = sendGet(`table/`+name+`?vde=true`, nil, &tableInfo)
	if err != nil {
		t.Error(err)
		return
	}
	if tableInfo.Filter != `ReadFilter`+name+`()` {
		t.Errorf(`wrong filter ` + tableInfo.Filter)
		return
	}

	err = sendPost(`content`, &url.Values{`vde`: {`true`}, `template`: {
		`DBFind(` + name + `, src).Limit(2)`}}, &retCont)
	if err != nil {
		t.Error(err)
		return
	}
	if !strings.Contains(RawToString(retCont.Tree), `No name`) {
		t.Errorf(`wrong tree %s`, RawToString(retCont.Tree))
		return
	}
}
