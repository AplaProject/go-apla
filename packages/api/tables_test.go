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
	"testing"
)

func TestTables(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	var ret tablesResult
	err := sendGet(`tables`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(`RET`, ret)
	if int64(ret.Count) < 7 {
		t.Error(fmt.Errorf(`The number of tables %d < 7`, ret.Count))
		return
	}
}

func TestTable(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	var ret tableResult
	err := sendGet(`table/keys`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if len(ret.Columns) == 0 {
		t.Error(err)
		return
	}
	err = sendGet(`table/contracts`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestJSONTable(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	name := randName(`json`)
	form := url.Values{"Name": {name}, "Columns": {`[{"name":"MyName","type":"varchar", "index": "0", 
	  "conditions":"true"}, {"name":"Doc", "type":"json","index": "0", "conditions":"true"}]`},
		"Permissions": {`{"insert": "true", "update" : "true", "new_column": "true"}`}}
	err := postTx(`NewTable`, &form)
	if err != nil {
		t.Error(err)
		return
	}

	form = url.Values{"Name": {name}, "Value": {`contract ` + name + ` {
		action { 
			var ret1, ret2 int
			ret1 = DBInsert("` + name + `", "MyName,Doc", "test", "{\"type\": \"0\"}")
			var mydoc map
			mydoc["type"] = "document"
			mydoc["ind"] = 2
			mydoc["doc"] = "Some text."
			ret2 = DBInsert("` + name + `", "MyName,Doc", "test2", mydoc)
		}}
		contract ` + name + `Upd {
		action {
			DBUpdate("` + name + `", 1, "Doc", "{\"type\": \"doc\", \"ind\": \"3\"}")
			var mydoc map
			mydoc["type"] = "doc"
			mydoc["doc"] = "Some test text."
			DBUpdate("` + name + `", 2, "myname,Doc", "test3", mydoc)
		}}
		`},
		"Conditions": {`ContractConditions("MainCondition")`}}
	err = postTx("NewContract", &form)
	if err != nil {
		t.Error(err)
		return
	}
	err = postTx(name, &url.Values{})
	if err != nil {
		t.Error(err)
		return
	}
	err = postTx(name+`Upd`, &url.Values{})
	if err != nil {
		t.Error(err)
		return
	}
}
