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
					if item.Params[0].Results[`error`] != err.Error() {
						t.Error(err)
						return
					}
					continue
				}
			} else {
				t.Error(err)
				return
			}
		}
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
	{`errTestMessage`, `contract errTestMessage {
		conditions {
		}
		action { qvar ivar int}
	}`,
		[]smartParams{
			{nil, map[string]string{`error`: `{"type":"panic","error":"unknown variable qvar"}`}},
		}},

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

func TestSignature(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	rnd := `rnd` + crypto.RandSeq(6)
	form := url.Values{`Value`: {`contract ` + rnd + `Transfer {
		    data {
				Recipient int
				Amount    money
				Signature string "optional hidden"
			}
			action { 
				$result = "OK " + Str($Amount)
				Println("TRANSFER CONTRACT", $result)
			}}
			
			contract ` + rnd + `Test {
				data {
					Recipient int "hidden"
					Amount  money
					Signature string "signature:` + rnd + `Transfer"
				}
				func action {
					` + rnd + `Transfer("Recipient,Amount,Signature",$Recipient,$Amount,$Signature )
					$result = "OOOPS " + Str($Amount)
					Println("TEST CONTRACT", $result)
				}
			  }
			`}, `Conditions`: {`true`}}
	if err := postTx(`NewContract`, &form); err != nil {
		t.Error(err)
		return
	}
	form = url.Values{`Name`: {rnd + `Transfer`}, `Value`: {`{"title": "Would you like to sign",
		"params":[
			{"name": "Receipient", "text": "Wallet"},
			{"name": "Amount", "text": "Amount(EGS)"}
			]}`}, `Conditions`: {`true`}}
	if err := postTx(`NewSign`, &form); err != nil {
		t.Error(err)
		return
	}
	err := postTx(rnd+`Test`, &url.Values{`Amount`: {`12345`}, `Recipient`: {`98765`}})
	if err != nil {
		t.Error(err)
		return
	}
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
		"tables": [
			{
				"Name": "members",
				"Columns": "[{\"name\":\"name\",\"type\":\"varchar\",\"conditions\":\"true\"},{\"name\":\"birthday\",\"type\":\"datetime\",\"conditions\":\"true\"},{\"name\":\"member_id\",\"type\":\"number\",\"conditions\":\"true\"},{\"name\":\"val\",\"type\":\"text\",\"conditions\":\"true\"},{\"name\":\"name_first\",\"type\":\"text\",\"conditions\":\"true\"},{\"name\":\"name_middle\",\"type\":\"text\",\"conditions\":\"true\"}]",
				"Permissions": "{\"insert\":\"true\",\"update\":\"true\",\"new_column\":\"true\"}"
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
	imppage = `{
	"pages": [
        {
            "Name": "profile_view",
            "Conditions": "ContractAccess(\"@1EditPage\")",
            "Menu": "default_menu",
            "Value": "Div(Class: content-wrapper){\r\n    Div(Class: content-heading, Body: LangRes(user_info))\r\n\r\n    If(#v_member_id# > 0){\r\n        DBFind(members, mysrc).Where(member_id=#v_member_id#).Vars(prefix)\r\n    }.Else{\r\n        DBFind(members, mysrc).Where(member_id=#key_id#).Vars(prefix)\r\n    }\r\n\r\n    If(#prefix_id#>0){\r\n    }.Else{\r\n        SetVar(prefix_username, \"\")\r\n        SetVar(prefix_name_last, \"\")\r\n        SetVar(prefix_name_first, \"\")\r\n        SetVar(prefix_name_middle, \"\")\r\n        SetVar(prefix_birthdate, \"01.01.1990\")\r\n    }\r\n\r\nDiv(row df f-valign){\r\n        Div(col-md-3)\r\n        Div(col-md-6){\r\n            Div(panel panel-default){\r\n                Form(){ \r\n\r\n\t\t\t\t\tDiv(list-group-item){\r\n\t\t\t\t\t\tSpan(Class: h3, Body: LangRes(user_info))\t\r\n\t\t\t\t\t}\r\n\r\n\t\t\t\t\tDiv(list-group-item){\r\n\t\t\t\t\t\tDiv(row df f-valign){\r\n\t\t\t\t\t\t\tDiv(col-md-12 mt-sm  text-center){\r\n\r\n\t\t\t\t\t\t\t\tIf(#prefix_id#>0){\r\n\t\t\t\t\t\t\t\t\tIf(#prefix_member_id# == #key_id#){\r\n\t\t\t\t\t\t\t\t\t\tButton(Class: btn btn-link, Page:profile_edit, PageParams:\"v_member_id=#member_id#\"){\r\n\t\t\t\t\t\t\t\t\t\t\tImage(\"#prefix_avatar#\",,img-circle).Style(width: 100px;  border: 1px solid #5A5D63; margin-bottom: 15px;)\r\n\t\t\t\t\t\t\t\t\t\t\tDiv(,Span(Class: h3 text-bold, Body: #prefix_username#))\r\n\t\t\t\t\t\t\t\t\t\t}\r\n\t\t\t\t\t\t\t\t\t}.Else{\r\n\t\t\t\t\t\t\t\t\t\tImage(\"#prefix_avatar#\",,img-circle).Style(width: 100px;  border: 1px solid #5A5D63; margin-bottom: 15px;)\r\n\t\t\t\t\t\t\t\t\t\tDiv(,Span(Class: h3 text-bold, Body: #prefix_username#))\r\n\t\t\t\t\t\t\t\t\t}\r\n\t\t\t\t\t\t\t\t}.Else{\r\n\t\t\t\t\t\t\t\t\tButton(Class: btn btn-link, Page:profile_edit){\r\n\t\t\t\t\t\t\t\t\t\tSpan(Class: h3 text-bold, Body: LangRes(editing_profile))\r\n\t\t\t\t\t\t\t\t\t}\r\n\t\t\t\t\t\t\t\t}\r\n\r\n\t\t\t\t\t\t\t}\r\n\t\t\t\t\t\t}\r\n\t\t\t\t\t}\r\n\r\n\t\t\t\t\tDiv(list-group-item){\r\n\t\t\t\t\t\tDiv(row df f-valign){\r\n\t\t\t\t\t\t\tDiv(col-md-6 mt-sm  text-right){\r\n\t\t\t\t\t\t\t\tSpan(Class: h4, Body: LangRes(name_last))\r\n\t\t\t\t\t\t\t}\r\n\t\t\t\t\t\t\tDiv(col-md-6 mt-sm text-left){\r\n\t\t\t\t\t\t\t\tSpan(Class: h4, Body: #prefix_name_last#)\r\n\t\t\t\t\t\t\t} \r\n\t\t\t\t\t\t}\r\n\t\t\t\t\t\tDiv(row df f-valign){\r\n\t\t\t\t\t\t\tDiv(col-md-6 mt-sm  text-right){\r\n\t\t\t\t\t\t\t\tSpan(Class: h4, Body: LangRes(name_first))\r\n\t\t\t\t\t\t\t}\r\n\t\t\t\t\t\t\tDiv(col-md-6 mt-sm text-left){\r\n\t\t\t\t\t\t\t\tSpan(Class: h4, Body: #prefix_name_first#)\r\n\t\t\t\t\t\t\t} \r\n\t\t\t\t\t\t}\r\n\t\t\t\t\t\tDiv(row df f-valign){\r\n\t\t\t\t\t\t\tDiv(col-md-6 mt-sm  text-right){\r\n\t\t\t\t\t\t\t\tSpan(Class: h4, Body: LangRes(name_middle))\r\n\t\t\t\t\t\t\t}\r\n\t\t\t\t\t\t\tDiv(col-md-6 mt-sm text-left){\r\n\t\t\t\t\t\t\t\tSpan(Class: h4, Body: #prefix_name_middle#)\r\n\t\t\t\t\t\t\t} \r\n\t\t\t\t\t\t}\r\n\t\t\t\t\t}\r\n\r\n\t\t\t\t\tDiv(list-group-item){\r\n\t\t\t\t\t\tDiv(row df f-valign){\r\n\t\t\t\t\t\t\tDiv(col-md-6 mt-sm  text-right){\r\n\t\t\t\t\t\t\t\tSpan(Class: h4, Body: LangRes(gender))\r\n\t\t\t\t\t\t\t}\r\n\t\t\t\t\t\t\tDiv(col-md-6 mt-sm text-left){\r\n\t\t\t\t\t\t\t\tSpan(Class: h4, Body: EcosysParam(gender_list, #prefix_gender#))\r\n\t\t\t\t\t\t\t} \r\n\t\t\t\t\t\t}\r\n\t\t\t\t\t}\r\n\r\n\t\t\t\t\tDiv(list-group-item){\r\n\t\t\t\t\t\tDiv(row df f-valign){\r\n\t\t\t\t\t\t\tDiv(col-md-6 mt-sm  text-right){\r\n\t\t\t\t\t\t\t\tSpan(Class: h4, Body: LangRes(birthdate))\r\n\t\t\t\t\t\t\t}\r\n\t\t\t\t\t\t\tDiv(col-md-6 mt-sm text-left){\r\n\t\t\t\t\t\t\t\tSpan(Class: h4, Body: DateTime(#prefix_birthdate#, \"DD.MM.YYYY\"))\r\n\t\t\t\t\t\t\t} \r\n\t\t\t\t\t\t}\r\n\t\t\t\t\t}\r\n\r\n\t\t\t\t\tDiv(list-group-item){\r\n\t\t\t\t\t\tDiv(row df f-valign){\r\n\t\t\t\t\t\t\tDiv(col-md-12 mt-sm  text-center){\r\n\t\t\t\t\t\t\t\tSpan(Class: h4 text-bold, Body: Address(#prefix_member_id#))\r\n\t\t\t\t\t\t\t\tDiv(,Span(Class: h5, Body: LangRes(member_id)))\r\n\t\t\t\t\t\t\t}\r\n\t\t\t\t\t\t}\r\n\t\t\t\t\t}\t\t\t\t\t\r\n\r\n                }\r\n            }\r\n        }\r\n        Div(col-md-3)\r\n    }\r\n}"
        }
	]
	}
`
)

func TestImport(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}

	form := url.Values{"Data": {imp}}
	_, _, err := postTxResult(`@1Import`, &form)
	if err != nil {
		t.Error(err)
		return
	}

}
