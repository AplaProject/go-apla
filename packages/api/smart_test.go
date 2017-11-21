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

type smartParams struct {
	Params  map[string]string
	Results map[string]string
}

type smartContract struct {
	Name   string
	Value  string
	Params []smartParams
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
		t.Error(fmt.Sprintf(`MainCondition name is wrong: %s`, cntResult.Name))
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
	if err := postTx(`MoneyTransfer`, &form); cutErr(err) != `Recipient 0005207000 is invalid` {
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
	id, msg, err := postTxResult(`NewParameter`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	err = postTx(`NewParameter`, &form)
	if cutErr(err) != fmt.Sprintf(`!Parameter %s already exists`, name) {
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
	if cutErr(err) != fmt.Sprintf(`!Menu %s already exists`, menuname) {
		t.Error(err)
		return
	}

	form = url.Values{"Name": {name + `23`}, "Value": {`New Param Value`},
		"Conditions": {`ContractConditions("MainCondition")`}}
	id, msg, err = postTxResult(`EditParameter`, &form)
	if cutErr(err) != fmt.Sprintf(`Record %s23 has not been found`, name) {
		t.Error(err)
		return
	}

	name = randName(`page`)
	form = url.Values{"Name": {name}, "Value": {value},
		"Menu": {menu}, "Conditions": {"ContractConditions(`MainCondition`)"}}
	id, msg, err = postTxResult(`NewPage`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	err = postTx(`NewPage`, &form)
	if cutErr(err) != fmt.Sprintf(`!Page %s already exists`, name) {
		t.Error(err)
		return
	}

	form = url.Values{"Name": {name}, "Value": {value},
		"Conditions": {"ContractConditions(`MainCondition`)"}}
	id, msg, err = postTxResult(`NewBlock`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	err = postTx(`NewBlock`, &form)
	if cutErr(err) != fmt.Sprintf(`!Block %s aready exists`, name) {
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
	id, msg, err = postTxResult(`EditPage`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"Id": {`1112`}, "Value": {value + `Span(Test)`},
		"Menu": {menu}, "Conditions": {"ContractConditions(`MainCondition`)"}}
	err = postTx(`EditPage`, &form)
	if cutErr(err) != `Item 1112 has not been found` {
		t.Error(err)
		return
	}

	form = url.Values{"Id": {`2`}, "Value": {`Span(Append)`}}
	id, msg, err = postTxResult(`AppendPage`, &form)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(`RET`, id, msg)
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
	{"name":"Active", "type":"character","index": "0", "conditions":"true"}]`},
		"Permissions": {`{"insert": "true", "update" : "true", "new_column": "true"}`}}
	err := postTx(`NewTable`, &form)
	if err != nil {
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
	form = url.Values{"TableName": {name}, "Name": {`newCol`},
		"Type": {"varchar"}, "Index": {"0"}, "Permissions": {"true"}}
	err = postTx(`NewColumn`, &form)
	if err != nil {
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
