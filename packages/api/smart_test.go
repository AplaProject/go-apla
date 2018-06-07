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

	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/shopspring/decimal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	{"name":"text", "type":"text","conditions":"true"}]`}, "ApplicationId": {`1`},
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
	}`}, `Conditions`: {`true`}, "ApplicationId": {`1`}}
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
	assert.NoError(t, keyLogin(1))

	name := randName(`page`)
	menuname := randName(`menu`)
	menu := `government`
	value := `P(test,test paragraph)`

	form := url.Values{"Name": {name}, "Value": {`Param Value`},
		"Conditions": {`ContractConditions("MainCondition")`}}
	assert.NoError(t, postTx(`NewParameter`, &form))

	err := postTx(`NewParameter`, &form)
	assert.Equal(t, fmt.Sprintf(`{"type":"warning","error":"Parameter %s already exists"}`, name), cutErr(err))

	form = url.Values{"Name": {menuname}, "Value": {`first
			second
			third`}, "Title": {`My Menu`},
		"Conditions": {`true`}}
	assert.NoError(t, postTx(`NewMenu`, &form))

	err = postTx(`NewMenu`, &form)
	assert.Equal(t, fmt.Sprintf(`{"type":"warning","error":"Menu %s already exists"}`, menuname), cutErr(err))

	form = url.Values{"Id": {`7123`}, "Value": {`New Param Value`},
		"Conditions": {`ContractConditions("MainCondition")`}}
	err = postTx(`EditParameter`, &form)
	assert.Equal(t, `{"type":"panic","error":"Item 7123 has not been found"}`, cutErr(err))

	form = url.Values{"Id": {`16`}, "Value": {`Changed Param Value`},
		"Conditions": {`ContractConditions("MainCondition")`}}
	assert.NoError(t, postTx(`EditParameter`, &form))

	name = randName(`page`)
	form = url.Values{"Name": {name}, "Value": {value}, "ApplicationId": {`1`},
		"Menu": {menu}, "Conditions": {"ContractConditions(`MainCondition`)"}}
	assert.NoError(t, postTx(`NewPage`, &form))

	err = postTx(`NewPage`, &form)
	assert.Equal(t, fmt.Sprintf(`{"type":"warning","error":"Page %s already exists"}`, name), cutErr(err))
	err = postTx(`NewPage`, &form)
	if cutErr(err) != fmt.Sprintf(`{"type":"warning","error":"Page %s already exists"}`, name) {
		t.Error(err)
		return
	}
	form = url.Values{"Name": {`app` + name}, "Value": {value}, "ValidateCount": {"2"},
		"ValidateMode": {"1"}, "ApplicationId": {`1`},
		"Menu": {menu}, "Conditions": {"ContractConditions(`MainCondition`)"}}
	err = postTx(`NewPage`, &form)
	if err != nil {
		t.Error(err)
		return
	}

	var ret listResult
	err = sendGet(`list/pages`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	id := ret.Count
	form = url.Values{"Id": {id}, "ValidateCount": {"2"}, "ValidateMode": {"1"}}
	err = postTx(`EditPage`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	var row rowResult
	err = sendGet(`row/pages/`+id, nil, &row)
	if err != nil {
		t.Error(err)
		return
	}

	if row.Value["validate_mode"] != `1` {
		t.Errorf(`wrong validate value %s`, row.Value["validate_mode"])
		return
	}

	form = url.Values{"Id": {id}, "Value": {value}, "ValidateCount": {"1"},
		"ValidateMode": {"0"}}
	err = postTx(`EditPage`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	err = sendGet(`row/pages/`+id, nil, &row)
	if err != nil {
		t.Error(err)
		return
	}
	if row.Value["validate_mode"] != `0` {
		t.Errorf(`wrong validate value %s`, row.Value["validate_mode"])
		return
	}

	form = url.Values{"Id": {id}, "Value": {value}, "ValidateCount": {"1"},
		"ValidateMode": {"0"}}
	err = postTx(`EditPage`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	err = sendGet(`row/pages/`+id, nil, &row)
	if err != nil {
		t.Error(err)
		return
	}
	if row.Value["validate_mode"] != `0` {
		t.Errorf(`wrong validate value %s`, row.Value["validate_mode"])
		return
	}

	form = url.Values{"Name": {name}, "Value": {value}, "ApplicationId": {`1`},
		"Conditions": {"ContractConditions(`MainCondition`)"}}
	assert.NoError(t, postTx(`NewBlock`, &form))

	err = postTx(`NewBlock`, &form)
	assert.EqualError(t, err, fmt.Sprintf(`{"type":"warning","error":"Block %s already exists"}`, name))

	form = url.Values{"Id": {`1`}, "Name": {name}, "Value": {value},
		"Conditions": {"ContractConditions(`MainCondition`)"}}
	assert.NoError(t, postTx(`EditBlock`, &form))

	form = url.Values{"Id": {`1`}, "Value": {value + `Span(Test)`},
		"Menu": {menu}, "Conditions": {"ContractConditions(`MainCondition`)"}}
	assert.NoError(t, postTx(`EditPage`, &form))

	form = url.Values{"Id": {`1112`}, "Value": {value + `Span(Test)`},
		"Menu": {menu}, "Conditions": {"ContractConditions(`MainCondition`)"}}
	err = postTx(`EditPage`, &form)
	assert.Equal(t, `{"type":"panic","error":"Item 1112 has not been found"}`, cutErr(err))

	form = url.Values{"Id": {`1`}, "Value": {`Span(Append)`}}
	assert.NoError(t, postTx(`AppendPage`, &form))
}

func TestNewTable(t *testing.T) {
	assert.NoError(t, keyLogin(1))

	name := randName(`tbl`)
	form := url.Values{"Name": {`1_` + name}, "ApplicationId": {"1"}, "Columns": {`[{"name":"MyName","type":"varchar", 
		"conditions":"true"},
	  {"name":"Name", "type":"varchar","index": "0", "conditions":"{\"read\":\"true\",\"update\":\"true\"}"}]`},
		"Permissions": {`{"insert": "true", "update" : "true", "new_column": "true"}`}}
	assert.NoError(t, postTx(`NewTable`, &form))

	form = url.Values{"TableName": {`1_` + name}, "Name": {`newCol`},
		"Type": {"varchar"}, "Index": {"0"}, "Permissions": {"true"}}
	assert.NoError(t, postTx(`NewColumn`, &form))

	form = url.Values{`Value`: {`contract sub` + name + ` {
		action {
			DBInsert("1_` + name + `", "name", "ok")
			DBUpdate("1_` + name + `", 1, "name", "test value" )
			$result = DBFind("1_` + name + `").Columns("name").WhereId(1).One("name")
		}
	}`}, `Conditions`: {`true`}, "ApplicationId": {"1"}}
	assert.NoError(t, postTx(`NewContract`, &form))

	_, msg, err := postTxResult(`sub`+name, &url.Values{})
	assert.NoError(t, err)
	assert.Equal(t, msg, "test value")

	form = url.Values{"Name": {name}, "ApplicationId": {"1"}, "Columns": {`[{"name":"MyName","type":"varchar", "index": "1", 
	  "conditions":"true"},
	{"name":"Amount", "type":"number","index": "0", "conditions":"true"},
	{"name":"Doc", "type":"json","index": "0", "conditions":"true"},	
	{"name":"Active", "type":"character","index": "0", "conditions":"true"}]`},
		"Permissions": {`{"insert": "true", "update" : "true", "new_column": "true"}`}}
	assert.NoError(t, postTx(`NewTable`, &form))

	assert.EqualError(t, postTx(`NewTable`, &form), fmt.Sprintf(`{"type":"panic","error":"Table %s exists"}`, name))

	form = url.Values{"Name": {name},
		"Permissions": {`{"insert": "ContractConditions(\"MainCondition\")",
				"update" : "true", "new_column": "ContractConditions(\"MainCondition\")"}`}}
	assert.NoError(t, postTx(`EditTable`, &form))

	form = url.Values{"TableName": {name}, "Name": {`newDoc`},
		"Type": {"json"}, "Index": {"0"}, "Permissions": {"true"}}
	assert.NoError(t, postTx(`NewColumn`, &form))

	form = url.Values{"TableName": {name}, "Name": {`newCol`},
		"Type": {"varchar"}, "Index": {"0"}, "Permissions": {"true"}}
	assert.NoError(t, postTx(`NewColumn`, &form))

	err = postTx(`NewColumn`, &form)
	if err.Error() != `{"type":"panic","error":"Column newcol exists"}` {
		t.Error(err)
		return
	}
	form = url.Values{"TableName": {name}, "Name": {`newCol`},
		"Permissions": {"ContractConditions(\"MainCondition\")"}}
	assert.NoError(t, postTx(`EditColumn`, &form))

	upname := strings.ToUpper(name)
	form = url.Values{"TableName": {upname}, "Name": {`UPCol`},
		"Type": {"varchar"}, "Index": {"0"}, "Permissions": {"true"}}
	assert.NoError(t, postTx(`NewColumn`, &form))

	form = url.Values{"TableName": {upname}, "Name": {`upCOL`},
		"Permissions": {"ContractConditions(\"MainCondition\")"}}
	assert.NoError(t, postTx(`EditColumn`, &form))

	form = url.Values{"Name": {upname},
		"Permissions": {`{"insert": "ContractConditions(\"MainCondition\")", 
			"update" : "true", "new_column": "ContractConditions(\"MainCondition\")"}`}}
	assert.NoError(t, postTx(`EditTable`, &form))

	var ret tablesResult
	assert.NoError(t, sendGet(`tables`, nil, &ret))
}

type invalidPar struct {
	Name  string
	Value string
}

func TestUpdateSysParam(t *testing.T) {
	assert.NoError(t, keyLogin(1))

	form := url.Values{"Name": {`max_columns`}, "Value": {`49`}}
	assert.NoError(t, postTx(`UpdateSysParam`, &form))

	var sysList ecosystemParamsResult
	assert.NoError(t, sendGet(`systemparams?names=max_columns`, nil, &sysList))
	assert.Len(t, sysList.List, 1)
	assert.Equal(t, "49", sysList.List[0].Value)

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
		"ApplicationId": {`1`}, "Conditions": {`ContractConditions("MainCondition")`}}
	assert.NoError(t, postTx("NewContract", &form))

	err := postTx(name, &form)
	if err != nil {
		assert.EqualError(t, err, `{"type":"panic","error":"Access denied"}`)
	}

	assert.NoError(t, sendGet(`systemparams?names=max_columns,max_indexes`, nil, &sysList))
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
		{`full_nodes`, `[["", "http://127.0.0.1", "100", "c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7"]]`},
		{`full_nodes`, `[["127.0.0.1", "", "100", "c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7c7d0"]]`},
		{`full_nodes`, `[["127.0.0.1", "http://127.0.0.1", "0", "c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7c7d0"]]`},
		{"full_nodes", "[]"},
	}
	for _, item := range notvalid {
		assert.Error(t, postTx(`UpdateSysParam`, &url.Values{`Name`: {item.Name}, `Value`: {item.Value}}))
		assert.NoError(t, sendGet(`systemparams?names=`+item.Name, nil, &sysList))
		assert.Len(t, sysList.List, 1, `have got wrong parameter `+item.Name)

		if len(sysList.List[0].Value) == 0 {
			continue
		}

		err = postTx(`UpdateSysParam`, &url.Values{`Name`: {item.Name}, `Value`: {sysList.List[0].Value}})
		assert.NoError(t, err, item.Name, sysList.List[0].Value, sysList.List[0])
	}
}

func TestUpdateFullNodesWithEmptyArray(t *testing.T) {
	require.NoErrorf(t, keyLogin(1), "on login")

	byteNodes := `[]`
	// byteNodes += `{"tcp_address":"127.0.0.1:7080", "api_address":"https://127.0.0.1:7081", "key_id":"5462687003324713865", "public_key":"4ea2433951ca21e6817426675874b2a6d98e5051c1100eddefa1847b0388e4834facf9abf427c46e2bc6cd5e3277fba533d03db553e499eb368194b3f1e514d4"}]`
	form := &url.Values{
		"Name":  {"full_nodes"},
		"Value": {string(byteNodes)},
	}

	require.EqualError(t, postTx(`UpdateSysParam`, form), `{"type":"panic","error":"Invalid value"}`)
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

func TestDBMetrics(t *testing.T) {
	assert.NoError(t, keyLogin(1))

	contract := randName("Metric")
	form := url.Values{
		"Value": {`
			contract ` + contract + ` {
				data {
					Metric string
				}
				conditions {}
				action {
					UpdateMetrics()
					$result = One(DBSelectMetrics($Metric, "1 days", "max"), "value")
				}
			}`},
		"Conditions": {"true"},
	}
	assert.NoError(t, postTx("NewContract", &form))

	metricValue := func(metric string) int {
		assert.NoError(t, postTx("UpdateMetrics", &url.Values{}))

		_, result, err := postTxResult(contract, &url.Values{"Metric": {metric}})
		assert.NoError(t, err)
		return converter.StrToInt(result)
	}

	ecosystemPages := metricValue("ecosystem_pages")
	ecosystemTx := metricValue("ecosystem_tx")

	form = url.Values{
		"Name":       {randName("page")},
		"Value":      {"P()"},
		"Menu":       {"default_menu"},
		"Conditions": {"true"},
	}
	assert.NoError(t, postTx("NewPage", &form))

	assert.Equal(t, 1, metricValue("ecosystem_pages")-ecosystemPages)
	assert.True(t, metricValue("ecosystem_tx") > ecosystemTx)

}

func TestPartitialEdit(t *testing.T) {
	assert.NoError(t, keyLogin(1))

	name := randName(`part`)
	form := url.Values{"Name": {name}, "Value": {"Span(Original text)"},
		"Menu": {"original_menu"}, "Conditions": {"ContractConditions(`MainCondition`)"}}
	assert.NoError(t, postTx(`NewPage`, &form))

	var retList listResult
	assert.NoError(t, sendGet(`list/pages`, nil, &retList))

	idItem := retList.Count
	value := `Span(Temp)`
	menu := `temp_menu`
	assert.NoError(t, postTx(`EditPage`, &url.Values{
		"Id":    {idItem},
		"Value": {value},
		"Menu":  {menu},
	}))

	var ret rowResult
	assert.NoError(t, sendGet(`row/pages/`+idItem, nil, &ret))
	assert.Equal(t, value, ret.Value["value"])
	assert.Equal(t, menu, ret.Value["menu"])

	value = `Span(Updated)`
	menu = `default_menu`
	conditions := `true`
	assert.NoError(t, postTx(`EditPage`, &url.Values{"Id": {idItem}, "Value": {value}}))
	assert.NoError(t, postTx(`EditPage`, &url.Values{"Id": {idItem}, "Menu": {menu}}))
	assert.NoError(t, postTx(`EditPage`, &url.Values{"Id": {idItem}, "Conditions": {conditions}}))
	assert.NoError(t, sendGet(`row/pages/`+idItem, nil, &ret))
	assert.Equal(t, value, ret.Value["value"])
	assert.Equal(t, menu, ret.Value["menu"])

	form = url.Values{"Name": {name}, "Value": {`MenuItem(One)`}, "Title": {`My Menu`},
		"Conditions": {"ContractConditions(`MainCondition`)"}}
	assert.NoError(t, postTx(`NewMenu`, &form))
	assert.NoError(t, sendGet(`list/menu`, nil, &retList))
	idItem = retList.Count
	value = `MenuItem(Two)`
	assert.NoError(t, postTx(`EditMenu`, &url.Values{"Id": {idItem}, "Value": {value}}))
	assert.NoError(t, postTx(`EditMenu`, &url.Values{"Id": {idItem}, "Conditions": {conditions}}))
	assert.NoError(t, sendGet(`row/menu/`+idItem, nil, &ret))
	assert.Equal(t, value, ret.Value["value"])
	assert.Equal(t, conditions, ret.Value["conditions"])

	form = url.Values{"Name": {name}, "Value": {`Span(Block)`},
		"Conditions": {"ContractConditions(`MainCondition`)"}}
	assert.NoError(t, postTx(`NewBlock`, &form))
	assert.NoError(t, sendGet(`list/blocks`, nil, &retList))
	idItem = retList.Count
	value = `Span(Updated block)`
	assert.NoError(t, postTx(`EditBlock`, &url.Values{"Id": {idItem}, "Value": {value}}))
	assert.NoError(t, postTx(`EditBlock`, &url.Values{"Id": {idItem}, "Conditions": {conditions}}))
	assert.NoError(t, sendGet(`row/blocks/`+idItem, nil, &ret))
	assert.Equal(t, value, ret.Value["value"])
	assert.Equal(t, conditions, ret.Value["conditions"])
}

func TestContractEdit(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	name := randName(`part`)
	form := url.Values{"Value": {`contract ` + name + ` {
		    action {
				$result = "before"
			}
		}`}, "ApplicationId": {`1`},
		"Conditions": {"ContractConditions(`MainCondition`)"}}
	err := postTx(`NewContract`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	var retList listResult
	err = sendGet(`list/contracts`, nil, &retList)
	if err != nil {
		t.Error(err)
		return
	}
	idItem := retList.Count
	value := `contract ` + name + ` {
		action {
			$result = "after"
		}
	}`
	conditions := `true`
	wallet := "1231234123412341230"
	err = postTx(`EditContract`, &url.Values{"Id": {idItem}, "Value": {value}})
	if err != nil {
		t.Error(err)
		return
	}
	err = postTx(`EditContract`, &url.Values{"Id": {idItem}, "Conditions": {conditions},
		"WalletId": {wallet}})
	if err != nil {
		t.Error(err)
		return
	}
	var ret rowResult
	err = sendGet(`row/contracts/`+idItem, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if ret.Value["value"] != value || ret.Value["conditions"] != conditions ||
		ret.Value["wallet_id"] != wallet {
		t.Errorf(`wrong parameters of contract`)
		return
	}
	_, msg, err := postTxResult(name, &url.Values{})
	if err != nil {
		t.Error(err)
		return
	}
	if msg != "after" {
		t.Errorf(`the wrong result of the contract %s`, msg)
	}
}

func TestDelayedContracts(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}

	form := url.Values{
		"Contract":   {"UnknownContract"},
		"EveryBlock": {"10"},
		"Limit":      {"2"},
		"Conditions": {"true"},
	}
	err := postTx("NewDelayedContract", &form)
	assert.EqualError(t, err, `{"type":"error","error":"Unknown contract @1UnknownContract"}`)

	form.Set("Contract", "MainCondition")
	err = postTx("NewDelayedContract", &form)
	assert.NoError(t, err)

	form.Set("BlockID", "1")
	err = postTx("NewDelayedContract", &form)
	assert.EqualError(t, err, `{"type":"error","error":"The blockID must be greater than the current blockID"}`)

	form = url.Values{
		"Id":         {"1"},
		"Contract":   {"MainCondition"},
		"EveryBlock": {"10"},
		"Conditions": {"true"},
		"Deleted":    {"1"},
	}
	err = postTx("EditDelayedContract", &form)
	assert.NoError(t, err)
}

func TestJSON(t *testing.T) {
	assert.NoError(t, keyLogin(1))

	contract := randName("JSONEncode")
	assert.NoError(t, postTx("NewContract", &url.Values{
		"Value": {`contract ` + contract + ` {
			action {
				var a array, m map
				m["k1"] = 1
				m["k2"] = 2
				a[0] = m
				a[1] = m

				info JSONEncode(a)
			}
		}`},
		"Conditions": {"true"},
	}))
	assert.EqualError(t, postTx(contract, &url.Values{}), `{"type":"info","error":"[{\"k1\":1,\"k2\":2},{\"k1\":1,\"k2\":2}]"}`)

	contract = randName("JSONDecode")
	assert.NoError(t, postTx("NewContract", &url.Values{
		"Value": {`contract ` + contract + ` {
			data {
				Input string
			}
			action {
				info Sprintf("%#v", JSONDecode($Input))
			}
		}`},
		"Conditions": {"true"},
	}))

	cases := []struct {
		source string
		result string
	}{
		{`"test"`, `{"type":"info","error":"\"test\""}`},
		{`["test"]`, `{"type":"info","error":"[]interface {}{\"test\"}"}`},
		{`{"test":1}`, `{"type":"info","error":"map[string]interface {}{\"test\":1}"}`},
		{`[{"test":1}]`, `{"type":"info","error":"[]interface {}{map[string]interface {}{\"test\":1}}"}`},
		{`{"test":1`, `{"type":"panic","error":"unexpected end of JSON input"}`},
	}

	for _, v := range cases {
		assert.EqualError(t, postTx(contract, &url.Values{"Input": {v.source}}), v.result)
	}
}

func TestBytesToString(t *testing.T) {
	assert.NoError(t, keyLogin(1))

	contract := randName("BytesToString")
	assert.NoError(t, postTx("NewContract", &url.Values{
		"Value": {`contract ` + contract + ` {
			data {
				File bytes "file"
			}
			action {
				$result = BytesToString($File)
			}
		}`},
		"Conditions": {"true"},
	}))

	content := crypto.RandSeq(100)
	_, res, err := postTxMultipart(contract, nil, map[string][]byte{"File": []byte(content)})
	assert.NoError(t, err)
	assert.Equal(t, content, res)
}

func TestMoneyDigits(t *testing.T) {
	assert.NoError(t, keyLogin(1))

	var v paramValue
	assert.NoError(t, sendGet("/ecosystemparam/money_digit", &url.Values{}, &v))

	contract := randName("MoneyDigits")
	assert.NoError(t, postTx("NewContract", &url.Values{
		"Value": {`contract ` + contract + ` {
			data {
				Value money
			}
			action {
				$result = $Value
			}
		}`},
		"ApplicationId": {"1"},
		"Conditions":    {"true"},
	}))

	_, result, err := postTxResult(contract, &url.Values{
		"Value": {"1"},
	})
	assert.NoError(t, err)

	d := decimal.New(1, int32(converter.StrToInt(v.Value)))
	assert.Equal(t, d.StringFixed(0), result)
}

func TestMemoryLimit(t *testing.T) {
	assert.NoError(t, keyLogin(1))

	contract := randName("Contract")
	assert.NoError(t, postTx("NewContract", &url.Values{
		"Value": {`contract ` + contract + ` {
			data {
				Count int "optional"
			}
			action {
				var a array
				while (true) {
					$Count = $Count + 1
					a[Len(a)] = JSONEncode(a)
				}
			}
		}`},
		"ApplicationId": {"1"},
		"Conditions":    {"true"},
	}))

	assert.EqualError(t, postTx(contract, &url.Values{}), `{"type":"panic","error":"Memory limit exceeded"}`)
}

func TestStack(t *testing.T) {
	assert.NoError(t, keyLogin(1))

	parent := randName("Parent")
	child := randName("Child")

	assert.NoError(t, postTx("NewContract", &url.Values{
		"Value": {`contract ` + child + ` {
			action {
				$result = $stack
			}
		}`},
		"ApplicationId": {"1"},
		"Conditions":    {"true"},
	}))

	assert.NoError(t, postTx("NewContract", &url.Values{
		"Value": {`contract ` + parent + ` {
			action {
				var arr array
				arr[0] = $stack
				arr[1] = ` + child + `()
				$result = arr
			}
		}`},
		"ApplicationId": {"1"},
		"Conditions":    {"true"},
	}))

	_, res, err := postTxResult(parent, &url.Values{})
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("[[@1%s] [@1%[1]s @1%s]]", parent, child), res)
}
