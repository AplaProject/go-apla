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

	"github.com/AplaProject/go-apla/packages/crypto"
)

type smartParams struct {
	Params  map[string]string
	Results map[string]string
}

type smartContract struct {
	Name   string
	Value  string
	Params []smartParams
}

func TestUpperName(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	rnd := crypto.RandSeq(4)
	form := url.Values{"Name": {"testTable" + rnd}, "Columns": {`[{"name":"num","type":"text",   "conditions":"true"},
	{"name":"text", "type":"text","conditions":"true"}]`},
		"Permissions": {`{"insert": "true", "update" : "true", "new_column": "true"}`}}
	err := postTx(`NewTable`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{`Value`: {`contract AddRow` + rnd + ` {
		data {
		}
		conditions {
		}
		action {
		   DBInsert("testTable` + rnd + `", "num, text", "fgdgf", "124234") 
		}
	}`}, `Conditions`: {`true`}}
	if err := postTx(`NewContract`, &form); err != nil {
		t.Error(err)
		return
	}
	if err := postTx(`AddRow`+rnd, &url.Values{}); err != nil {
		t.Error(err)
		return
	}
}

func TestSmartFields(t *testing.T) {

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	var cntResult getContractResult
	err := sendGet(`contract/MainCondition`, nil, &cntResult)
	if err != nil {
		t.Error(err)
		return
	}
	if len(cntResult.Fields) != 0 {
		t.Error(`MainCondition fields must be empty`)
		return
	}
	if cntResult.Name != `@1MainCondition` {
		t.Errorf(`MainCondition name is wrong: %s`, cntResult.Name)
		return
	}
	if err := postTx(`MainCondition`, &url.Values{}); err != nil {
		t.Error(err)
		return
	}
}

func TestMoneyTransfer(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}

	form := url.Values{`Amount`: {`53330000`}, `Recipient`: {`0005-2070-2000-0006-0200`}}
	if err := postTx(`MoneyTransfer`, &form); err != nil {
		t.Error(err)
		return
	}
	form = url.Values{`Amount`: {`2440000`}, `Recipient`: {`1109-7770-3360-6764-7059`}, `Comment`: {`Test`}}
	if err := postTx(`MoneyTransfer`, &form); err != nil {
		t.Error(err)
		return
	}
	form = url.Values{`Amount`: {`53330000`}, `Recipient`: {`0005207000`}}
	if err := postTx(`MoneyTransfer`, &form); cutErr(err) != `{"type":"error","error":"Recipient 0005207000 is invalid"}` {
		t.Error(err)
		return
	}
}

func TestPage(t *testing.T) {

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	name := randName(`page`)
	menuname := randName(`menu`)
	menu := `government`
	value := `P(test,test paragraph)`

	form := url.Values{"Name": {name}, "Value": {`Param Value`},
		"Conditions": {`ContractConditions("MainCondition")`}}
	err := postTx(`NewParameter`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	err = postTx(`NewParameter`, &form)
	if cutErr(err) != fmt.Sprintf(`{"type":"warning","error":"Parameter %s already exists"}`, name) {
		t.Error(err)
		return
	}

	form = url.Values{"Name": {menuname}, "Value": {`first
			second
			third`}, "Title": {`My Menu`},
		"Conditions": {`true`}}
	err = postTx(`NewMenu`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	err = postTx(`NewMenu`, &form)
	if cutErr(err) != fmt.Sprintf(`{"type":"warning","error":"Menu %s already exists"}`, menuname) {
		t.Error(err)
		return
	}

	form = url.Values{"Id": {`7123`}, "Value": {`New Param Value`},
		"Conditions": {`ContractConditions("MainCondition")`}}
	err = postTx(`EditParameter`, &form)
	if cutErr(err) != `{"type":"panic","error":"Item 7123 has not been found"}` {
		t.Error(err)
		return
	}
	form = url.Values{"Id": {`13`}, "Value": {`Changed Param Value`},
		"Conditions": {`ContractConditions("MainCondition")`}}
	err = postTx(`EditParameter`, &form)
	if err != nil {
		t.Error(err)
		return
	}

	name = randName(`page`)
	form = url.Values{"Name": {name}, "Value": {value},
		"Menu": {menu}, "Conditions": {"ContractConditions(`MainCondition`)"}}
	err = postTx(`NewPage`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	err = postTx(`NewPage`, &form)
	if cutErr(err) != fmt.Sprintf(`{"type":"warning","error":"Page %s already exists"}`, name) {
		t.Error(err)
		return
	}

	form = url.Values{"Name": {name}, "Value": {value},
		"Conditions": {"ContractConditions(`MainCondition`)"}}
	err = postTx(`NewBlock`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	err = postTx(`NewBlock`, &form)
	if err.Error() != fmt.Sprintf(`{"type":"warning","error":"Block %s already exists"}`, name) {
		t.Error(err)
		return
	}
	form = url.Values{"Id": {`1`}, "Name": {name}, "Value": {value},
		"Conditions": {"ContractConditions(`MainCondition`)"}}
	err = postTx(`EditBlock`, &form)
	if err != nil {
		t.Error(err)
		return
	}

	form = url.Values{"Id": {`1`}, "Value": {value + `Span(Test)`},
		"Menu": {menu}, "Conditions": {"ContractConditions(`MainCondition`)"}}
	err = postTx(`EditPage`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"Id": {`1112`}, "Value": {value + `Span(Test)`},
		"Menu": {menu}, "Conditions": {"ContractConditions(`MainCondition`)"}}
	err = postTx(`EditPage`, &form)
	if cutErr(err) != `{"type":"panic","error":"Item 1112 has not been found"}` {
		t.Error(err)
		return
	}

	form = url.Values{"Id": {`1`}, "Value": {`Span(Append)`}}
	err = postTx(`AppendPage`, &form)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestNewTable(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	name := randName(`tbl`)
	form := url.Values{"Name": {name}, "Columns": {`[{"name":"MyName","type":"varchar", "index": "1", 
	  "conditions":"true"},
	{"name":"Amount", "type":"number","index": "0", "conditions":"true"},
	{"name":"Doc", "type":"json","index": "0", "conditions":"true"},
	{"name":"Active", "type":"character","index": "0", "conditions":"true"}]`},
		"Permissions": {`{"insert": "true", "update" : "true", "new_column": "true"}`}}
	err := postTx(`NewTable`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	err = postTx(`NewTable`, &form)
	if err.Error() != fmt.Sprintf(`{"type":"panic","error":"table %s exists"}`, name) {
		t.Error(err)
		return
	}
	form = url.Values{"Name": {name},
		"Permissions": {`{"insert": "ContractConditions(\"MainCondition\")", 
			"update" : "true", "new_column": "ContractConditions(\"MainCondition\")"}`}}
	err = postTx(`EditTable`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"TableName": {name}, "Name": {`newDoc`},
		"Type": {"json"}, "Index": {"0"}, "Permissions": {"true"}}
	err = postTx(`NewColumn`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"TableName": {name}, "Name": {`newCol`},
		"Type": {"varchar"}, "Index": {"0"}, "Permissions": {"true"}}
	err = postTx(`NewColumn`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	err = postTx(`NewColumn`, &form)
	if err.Error() != `{"type":"panic","error":"column newcol exists"}` {
		t.Error(err)
		return
	}
	form = url.Values{"TableName": {name}, "Name": {`newCol`},
		"Permissions": {"ContractConditions(\"MainCondition\")"}}
	err = postTx(`EditColumn`, &form)
	if err != nil {
		t.Error(err)
		return
	}
}

type invalidPar struct {
	Name  string
	Value string
}

func TestUpdateSysParam(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	form := url.Values{"Name": {`max_columns`}, "Value": {`49`}}
	err := postTx(`UpdateSysParam`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	var sysList ecosystemParamsResult
	err = sendGet(`systemparams?names=max_columns`, nil, &sysList)
	if err != nil {
		t.Error(err)
		return
	}
	if len(sysList.List) != 1 || sysList.List[0].Value != `49` {
		t.Error(`Wrong max_column value`)
		return
	}
	name := randName(`test`)
	form = url.Values{"Name": {name}, "Value": {`contract ` + name + ` {
		action { 
			var costlen int
			costlen = SysParamInt("extend_cost_len") + 1
			UpdateSysParam("Name,Value","max_columns","51")
			DBUpdateSysParam("extend_cost_len", Str(costlen), "true" )
			if SysParamInt("extend_cost_len") != costlen {
				error "Incorrect updated value"
			}
			DBUpdateSysParam("max_indexes", "4", "false" )
		}
		}`},
		"Conditions": {`ContractConditions("MainCondition")`}}
	err = postTx("NewContract", &form)
	if err != nil {
		t.Error(err)
		return
	}
	err = postTx(name, &form)
	if err != nil {
		if err.Error() != `{"type":"panic","error":"Access denied"}` {
			t.Error(err)
			return
		}
	}
	err = sendGet(`systemparams?names=max_columns,max_indexes`, nil, &sysList)
	if err != nil {
		t.Error(err)
		return
	}
	if len(sysList.List) != 2 || !((sysList.List[0].Value == `51` && sysList.List[1].Value == `4`) ||
		(sysList.List[0].Value == `4` && sysList.List[1].Value == `51`)) {
		t.Error(`Wrong max_column or max_indexes value`)
		return
	}
	err = postTx(name, &form)
	if err == nil || err.Error() != `{"type":"panic","error":"Access denied"}` {
		t.Error(`incorrect access to system parameter`)
		return
	}
	notvalid := []invalidPar{
		{`gap_between_blocks`, `100000`},
		{`rb_blocks_1`, `-1`},
		{`page_price`, `-20`},
		{`max_block_size`, `0`},
		{`max_fuel_tx`, `20string`},
		{`fuel_rate`, `string`},
		{`fuel_rate`, `[test]`},
		{`fuel_rate`, `[["name", "100"]]`},
		{`commission_wallet`, `[["1", "0"]]`},
		{`commission_wallet`, `[{"1", "50"}]`},
		{`full_nodes`, `[["34.12.25", "10", "c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7c7d0"]]`},
		{`full_nodes`, `[["1.34.12.25", "100", "c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7"]]`},
		{`full_nodes`, `[["34.12.25.100:65d321", "100000000000", "c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7c7d0"]]`},
	}
	for _, item := range notvalid {
		err = postTx(`UpdateSysParam`, &url.Values{`Name`: {item.Name}, `Value`: {item.Value}})
		if err == nil {
			t.Error(`must be invalid ` + item.Value)
			return
		}
		err = sendGet(`systemparams?names=`+item.Name, nil, &sysList)
		if err != nil {
			t.Error(err)
			return
		}
		if len(sysList.List) != 1 {
			t.Error(`have got wrong parameter ` + item.Name)
			return
		}
		err = postTx(`UpdateSysParam`, &url.Values{`Name`: {item.Name}, `Value`: {sysList.List[0].Value}})
		if err != nil {
			fmt.Println(item.Name, sysList.List[0].Value, sysList.List[0])
			t.Error(err)
			return
		}
	}
}

func TestValidateConditions(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}

	baseForm := url.Values{"Id": {"1"}, "Value": {"Test"}, "Conditions": {"incorrectConditions"}}
	contracts := map[string]url.Values{
		"EditContract":  baseForm,
		"EditParameter": baseForm,
		"EditMenu":      baseForm,
		"EditPage":      url.Values{"Id": {"1"}, "Value": {"Test"}, "Conditions": {"incorrectConditions"}, "Menu": {"1"}},
	}
	expectedErr := `{"type":"panic","error":"unknown identifier incorrectConditions"}`

	for contract, form := range contracts {
		err := postTx(contract, &form)
		if err.Error() != expectedErr {
			t.Errorf("contract %s expected '%s' got '%s'", contract, expectedErr, err)
			return
		}
	}
}
