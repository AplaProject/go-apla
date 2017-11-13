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

package apiv2

import (
	//	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/AplaProject/go-apla/packages/crypto"
)

func TestNewContracts(t *testing.T) {

	wanted := func(name, want string) bool {
		var ret getTestResult
		err := sendPost(`test/`+name, nil, &ret)
		if err != nil {
			t.Error(err)
			return false
		}
		if ret.Value != want {
			t.Error(fmt.Errorf(`%s != %s`, ret.Value, want))
			return false
		}
		return true
	}

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	for _, item := range contracts {
		var ret getContractResult
		err := sendGet(`contract/`+item.Name, nil, &ret)
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(errors[`E_CONTRACT`], item.Name)) {
				form := url.Values{"Name": {item.Name}, "Value": {item.Value},
					"Conditions": {`true`}}
				if err := postTx(`NewContract`, &form); err != nil {
					if item.Name != `errTest` || !strings.HasPrefix(err.Error(), `must be type 7d01 125 [Ln:4 Col:22]`) {
						t.Error(err)
						return
					}
				}
			} else {
				t.Error(err)
				return
			}
		}
		/*if strings.HasPrefix(item.Name, `EditProfile`) {
			form := url.Values{"Id": {`62`}, "Value": {item.Value},
				"Conditions": {`true`}}
			if err := postTx(`EditContract`, &form); err != nil {
				t.Error(err)
				return
			}
		}*/
		if strings.HasSuffix(item.Name, `testUpd`) {
			continue
		}
		for _, par := range item.Params {
			form := url.Values{}
			for key, value := range par.Params {
				form[key] = []string{value}
			}
			if err := postTx(item.Name, &form); err != nil {
				t.Error(err)
				return
			}
			for key, value := range par.Results {
				if !wanted(key, value) {
					return
				}
			}
		}
	}
}

var contracts = []smartContract{
	{`EditProfile6`, `contract EditProfile6 {
		data {
		}
		conditions {
		}
		action {
			$ret = DBFind("contracts").Columns("id,value").Where("id>= ? and id<= ?",3,5).Order("id")
			Test("edit",  "edit value 0")
		}
	}`,
		[]smartParams{
			{nil, map[string]string{`edit`: `edit value 0`}},
		}},

	{`TestDBFindOK`, `
		contract TestDBFindOK {
		action {
			var ret array
			var vals map
			ret = DBFind("contracts").Columns("id,value").Where("id>= ? and id<= ?",3,5).Order("id")
			if Len(ret) {
				Test("0",  "1")	
			} else {
				Test("0",  "0")	
			}
			ret = DBFind("contracts").Limit(3)
			if Len(ret) == 3 {
				Test("1",  "1")	
			} else {
				Test("1",  "0")	
			}
			ret = DBFind("contracts").Order("id").Offset(1).Limit(1)
			if Len(ret) != 1 {
				Test("2",  "0")	
			} else {
				vals = ret[0]
				Test("2",  vals["id"])	
			}
			ret = DBFind("contracts").Columns("id,rb_id").Order("id").Offset(1).Limit(1)
			if Len(ret) != 1 {
				Test("3",  "0")	
			} else {
				vals = ret[0]
				Test("3", vals["value"] + vals["id"])	
			}
			ret = DBFind("contracts").Columns("id,rb_id").Where("id='1'")
			if Len(ret) != 1 {
				Test("4",  "0")	
			} else {
				vals = ret[0]
				Test("4", vals["id"])	
			}
			ret = DBFind("contracts").Columns("id,rb_id").Where("id='1'")
			if Len(ret) != 1 {
				Test("4",  "0")	
			} else {
				vals = ret[0]
				Test("4", vals["id"])	
			}
			ret = DBFind("contracts").Columns("id,value").Where("id> ? and id < ?", 3, 8).Order("id")
			if Len(ret) != 4 {
				Test("5",  "0")	
			} else {
				vals = ret[0]
				Test("5", vals["id"])	
			}
			ret = DBFind("contracts").WhereId(7)
			if Len(ret) != 1 {
				Test("6",  "0")	
			} else {
				vals = ret[0]
				Test("6", vals["id"])	
			}
			Test("255",  "255")	
		}
	}`,
		[]smartParams{
			{nil, map[string]string{`0`: `1`, `1`: `1`, `2`: `2`, `3`: `2`, `4`: `1`, `5`: `4`,
				`6`:   `7`,
				`255`: `255`}},
		}},
	{`testEmpty`, `contract testEmpty {
				action { Test("empty",  "empty value")}}`,
		[]smartParams{
			{nil, map[string]string{`empty`: `empty value`}},
		}},
	{`testUpd`, `contract testUpd {
					action { Test("date",  "-2006.01.02-")}}`,
		[]smartParams{
			{nil, map[string]string{`date`: `-` + time.Now().Format(`2006.01.02`) + `-`}},
		}},
	{`testSimple`, `contract testSimple {
				data {
					amount int
					name   string
				}
				conditions {
					Test("scond", $amount, $name)
				}
				action { Test("sact", $name, $amount)}}`,
		[]smartParams{
			{map[string]string{`name`: `Simple name`, `amount`: `-56781`},
				map[string]string{`scond`: `-56781Simple name`,
					`sact`: `Simple name-56781`}},
		}},
	{`errTestVar`, `contract errTestVar {
			conditions {
			}
			action { var ivar int}
		}`,
		nil},
}

func TestEditContracts(t *testing.T) {

	wanted := func(name, want string) bool {
		var ret getTestResult
		err := sendPost(`test/`+name, nil, &ret)
		if err != nil {
			t.Error(err)
			return false
		}
		if ret.Value != want {
			t.Error(fmt.Errorf(`%s != %s`, ret.Value, want))
			return false
		}
		return true
	}

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	var cntlist contractsResult
	err := sendGet(`contracts`, nil, &cntlist)
	if err != nil {
		t.Error(err)
		return
	}
	var ret getContractResult
	err = sendGet(`contract/testUpd`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	sid := ret.TableID
	var row rowResult
	err = sendGet(`row/contracts/`+sid, nil, &row)
	if err != nil {
		t.Error(err)
		return
	}
	code := row.Value[`value`]
	off := strings.IndexByte(code, '-')
	newCode := code[:off+1] + time.Now().Format(`2006.01.02`) + code[off+11:]
	form := url.Values{`Id`: {sid}, `Value`: {newCode}, `Conditions`: {row.Value[`conditions`]}}
	if err := postTx(`EditContract`, &form); err != nil {
		t.Error(err)
		return
	}

	for _, item := range contracts {
		if !strings.HasSuffix(item.Name, `testUpd`) {
			continue
		}
		for _, par := range item.Params {
			form := url.Values{}
			for key, value := range par.Params {
				form[key] = []string{value}
			}
			if err := postTx(item.Name, &form); err != nil {
				t.Error(err)
				return
			}
			for key, value := range par.Results {
				if !wanted(key, value) {
					return
				}
			}
		}
	}
}

func TestActivateContracts(t *testing.T) {

	wanted := func(name, want string) bool {
		var ret getTestResult
		err := sendPost(`test/`+name, nil, &ret)
		if err != nil {
			t.Error(err)
			return false
		}
		if ret.Value != want {
			t.Error(fmt.Errorf(`%s != %s`, ret.Value, want))
			return false
		}
		return true
	}

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	rnd := `rnd` + crypto.RandSeq(6)
	form := url.Values{`Value`: {`contract ` + rnd + ` {
		    data {
				Par string
			}
			action { Test("active",  $Par)}}`}, `Conditions`: {`true`}}
	if err := postTx(`NewContract`, &form); err != nil {
		t.Error(err)
		return
	}
	var ret getContractResult
	err := sendGet(`contract/`+rnd, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if err := postTx(`ActivateContract`, &url.Values{`Id`: {ret.TableID}}); err != nil {
		t.Error(err)
		return
	}
	err = sendGet(`contract/`+rnd, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if !ret.Active {
		t.Error(fmt.Errorf(`Not activate ` + rnd))
	}
	var row rowResult
	err = sendGet(`row/contracts/`+ret.TableID, nil, &row)
	if err != nil {
		t.Error(err)
		return
	}
	if row.Value[`active`] != `1` {
		t.Error(fmt.Errorf(`row not activate ` + rnd))
	}

	if err := postTx(rnd, &url.Values{`Par`: {rnd}}); err != nil {
		t.Error(err)
		return
	}
	if !wanted(`active`, rnd) {
		return
	}
}

func TestContracts(t *testing.T) {

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}

	var ret contractsResult
	err := sendGet(`contracts`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(`RET`, ret)
}

var (
	impcont = `{
		 "contracts": [
        {
            "Name": "EditColumn_IMPORTED2",
            "Value": "contract EditColumn_IMPORTED2 {\n    data {\n    \tTableName   string\n\t    Name        string\n\t    Permissions string\n    }\n    conditions {\n        ColumnCondition($TableName, $Name, \"\", $Permissions, \"\")\n    }\n    action {\n        PermColumn($TableName, $Name, $Permissions)\n    }\n}",
            "Conditions": "ContractConditions(` + "`MainCondition`" + `)"
		}
	]
}`
	imp = `{
		"menus": [
			{
				"Name": "test_menu0679",
				"Conditions": "ContractAccess(\"@1EditMenu\")",
				"Value": "MenuItem(main, Default Ecosystem Menu)"
			}
		],
		"pages": [
			{
				"Name": "test_page05679",
				"Conditions": "ContractAccess(\"@1EditPage\")",
				"Menu": "default_menu",
				"Value": "P(class, Default Ecosystem Page)\nImage().Style(width:100px;)"
			}
		],
		"blocks": [
			{
				"Name": "test_block05679",
				"Conditions": "true",
				"Value": "block content"
			},
			{
				"Name": "test_block056790",
				"Conditions": "true",
				"Value": "block content"
			},
			{
				"Name": "test_block056791",
				"Conditions": "true",
				"Value": "block content"
			}
		],
		"parameters": [
			{
				"Name": "host01345679",
				"Value": "",
				"Conditions": "ContractConditions(` + "`MainCondition`" + `)"
			},
			{
				"Name": "host091",
				"Value": "Русский текст",
				"Conditions": "ContractConditions(` + "`MainCondition`" + `)"
			}
		]
	}`
	/*
			"tables": [
			{
				"Name": "members",
				"Columns": "[{\"name\":\"name\",\"type\":\"varchar\",\"conditions\":\"true\"},{\"name\":\"birthday\",\"type\":\"datetime\",\"conditions\":\"true\"},{\"name\":\"member_id\",\"type\":\"number\",\"conditions\":\"true\"},{\"name\":\"val\",\"type\":\"text\",\"conditions\":\"true\"},{\"name\":\"name_first\",\"type\":\"text\",\"conditions\":\"true\"},{\"name\":\"name_middle\",\"type\":\"text\",\"conditions\":\"true\"}]",
				"Permissions": "{\"insert\":\"true\",\"update\":\"true\",\"new_column\":\"true\"}"
			}
		],

	*/
	impdata = `{
		"data": [
	   {
		   "Table": "members",
		   "Columns": ["name","val"],
		   "Data": [
			   ["Bob","Richard mark"],
			   ["Mike Winter","Alan summer"]
			]
	   }
   ]
}`
)

func TestImport(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}

	form := url.Values{"Data": {impdata}}
	id, msg, err := postTxResult(`@1Import`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(`Import`, id, msg)
	t.Error(`OK`)
}
