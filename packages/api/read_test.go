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
	//	"strings"
	"testing"
)

func TestRead(t *testing.T) {

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	name := randName(`tbl`)
	form := url.Values{"Name": {name}, "Columns": {`[{"name":"my","type":"varchar", "index": "1", 
	  "conditions":"true"},
	{"name":"amount", "type":"number","index": "0", "conditions":"true"},
	{"name":"active", "type":"character","index": "0", "conditions":"true"}]`},
		"Permissions": {`{"insert": "true", "update" : "true", "read": "true", "new_column": "true"}`}}
	err := postTx(`NewTable`, &form)
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
			Println("GET Start")
			row = DBFind("%[1]s").Where("id>= ? and id<= ?", 2, 5)
			Println("ROW", row)
		}
	}

	contract GetData%[1]s {
		action {
			var row array
			Println("GET Start")
			row = DBFind("%[1]s").Columns("my").Where("id>= ? and id<= ?", 2, 5)
			Println("ROW", row)
		}
	}

	contract MyRead%[1]s {
		conditions {
			Println("MYREAD", $key_id)
			Println("MYREAD=", $table)
			if $access == "read" {
				var i int
				while i < Len($columns) {
					if $columns[i] == "*" || $columns[i] == "amount" {
						error "Access denied to amount"
					}
					i = i + 1
				}
		    }
		}
	}
	`, name)
	form = url.Values{"Value": {contFill},
		"Conditions": {`true`}}
	if err := postTx(`NewContract`, &form); err != nil {
		t.Error(err)
		return
	}
	if err := postTx(name, &url.Values{}); err != nil {
		t.Error(err)
		return
	}

	if err := postTx(`Get`+name, &url.Values{}); err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"Name": {name},
		"Permissions": {`{"insert": "ContractConditions(\"MainCondition\")", 
			"update" : "true", 
			"read": "ContractConditions(\"MyRead` + name + `\")",
			"new_column": "ContractConditions(\"MainCondition\")"}`}}
	err = postTx(`EditTable`, &form)
	if err != nil {
		t.Error(err)
		return
	}

	if err := postTx(`Get`+name, &url.Values{}); err != nil {
		t.Error(err)
		return
	}

	t.Error(`OK`)
}
