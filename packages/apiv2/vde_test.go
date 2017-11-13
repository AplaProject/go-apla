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
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
)

func TestVDECreate(t *testing.T) {
	var (
		err   error
		retid int64
		ret   vdeCreateResult
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

	rnd := `rnd` + crypto.RandSeq(6)
	form := url.Values{`Value`: {`contract ` + rnd + ` {
		    data {
				Par string
			}
			action { Test("active",  $Par)}}`}, `Conditions`: {`ContractConditions("MainCondition")`}, `vde`: {`true`}}

	if retid, _, err = postTxResult(`NewContract`, &form); err != nil {
		t.Error(err)
		return
	}
	form = url.Values{`Id`: {converter.Int64ToStr(retid)}, `Value`: {`contract ` + rnd + ` {
		data {
			Par string
		}
		action { Test("active 5",  $Par)}}`}, `Conditions`: {`ContractConditions("MainCondition")`}, `vde`: {`true`}}

	if err := postTx(`EditContract`, &form); err != nil {
		t.Error(err)
		return
	}
	form = url.Values{`Name`: {rnd}, `Value`: {`Test value`}, `Conditions`: {`ContractConditions("MainCondition")`},
		`vde`: {`1`}}

	if retid, _, err = postTxResult(`NewParameter`, &form); err != nil {
		t.Error(err)
		return
	}
	form = url.Values{`Name`: {`new_table`}, `Value`: {`Test value`}, `Conditions`: {`ContractConditions("MainCondition")`},
		`vde`: {`1`}}
	if err = postTx(`NewParameter`, &form); err != nil && err.Error() !=
		`500 {"error": "E_SERVER", "msg": "!Parameter new_table already exists" }` {
		t.Error(err)
		return
	}
	form = url.Values{`Id`: {converter.Int64ToStr(retid)}, `Value`: {`Test edit value`}, `Conditions`: {`true`},
		`vde`: {`1`}}
	if retid, _, err = postTxResult(`EditParameter`, &form); err != nil {
		t.Error(err)
		return
	}

	form = url.Values{"Name": {`menu` + rnd}, "Value": {`first
		second
		third`}, "Title": {`My Menu`},
		"Conditions": {`true`}, `vde`: {`1`}}
	retid, _, err = postTxResult(`NewMenu`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{`Id`: {converter.Int64ToStr(retid)}, `Value`: {`Test edit value`},
		`Conditions`: {`true`},
		`vde`:        {`1`}}
	if err = postTx(`EditMenu`, &form); err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"Id": {converter.Int64ToStr(retid)}, "Value": {`Span(Append)`},
		`vde`: {`1`}}
	err = postTx(`AppendMenu`, &form)
	if err != nil {
		t.Error(err)
		return
	}

	form = url.Values{"Name": {`page` + rnd}, "Value": {`Page`}, "Menu": {`government`},
		"Conditions": {`true`}, `vde`: {`1`}}
	retid, _, err = postTxResult(`NewPage`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{`Id`: {converter.Int64ToStr(retid)}, `Value`: {`Test edit page value`},
		`Conditions`: {`true`}, "Menu": {`government`},
		`vde`: {`1`}}
	if err = postTx(`EditPage`, &form); err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"Id": {converter.Int64ToStr(retid)}, "Value": {`Span(Test Page)`},
		`vde`: {`1`}}
	err = postTx(`AppendPage`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"Name": {`block` + rnd}, "Value": {`Page block`}, "Conditions": {`true`}, `vde`: {`1`}}
	retid, _, err = postTxResult(`NewBlock`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{`Id`: {converter.Int64ToStr(retid)}, `Value`: {`Test edit block value`},
		`Conditions`: {`true`}, `vde`: {`1`}}
	if err = postTx(`EditBlock`, &form); err != nil {
		t.Error(err)
		return
	}

	name := randName(`tbl`)
	form = url.Values{"Name": {name}, `vde`: {`1`}, "Columns": {`[{"name":"MyName","type":"varchar", "index": "1",
			  "conditions":"true"},
			{"name":"Amount", "type":"number","index": "0", "conditions":"true"},
			{"name":"Active", "type":"character","index": "0", "conditions":"true"}]`},
		"Permissions": {`{"insert": "true", "update" : "true", "new_column": "true"}`}}
	err = postTx(`NewTable`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"Name": {name}, `vde`: {`1`},
		"Permissions": {`{"insert": "ContractConditions(\"MainCondition\")",
						"update" : "true", "new_column": "ContractConditions(\"MainCondition\")"}`}}
	err = postTx(`EditTable`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"TableName": {name}, "Name": {`newCol`}, `vde`: {`1`},
		"Type": {"varchar"}, "Index": {"0"}, "Permissions": {"true"}}
	err = postTx(`NewColumn`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"TableName": {name}, "Name": {`newCol`}, `vde`: {`1`},
		"Permissions": {"ContractConditions(\"MainCondition\")"}}
	err = postTx(`EditColumn`, &form)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(`OK`)
}

func TestVDEParams(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	rnd := `rnd` + crypto.RandSeq(6)
	form := url.Values{`Name`: {rnd}, `Value`: {`Test value`}, `Conditions`: {`ContractConditions("MainCondition")`},
		`vde`: {`true`}}
	if _, _, err := postTxResult(`NewParameter`, &form); err != nil {
		t.Error(err)
		return
	}

	var ret ecosystemParamsResult
	err := sendGet(`ecosystemparams?vde=true`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if len(ret.List) < 5 {
		t.Error(fmt.Errorf(`wrong count of parameters %d`, len(ret.List)))
	}
	err = sendGet(`ecosystemparams?vde=true&names=stylesheet,`+rnd, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if len(ret.List) != 2 {
		t.Error(fmt.Errorf(`wrong count of parameters %d`, len(ret.List)))
	}
	var parValue paramValue
	err = sendGet(`ecosystemparam/`+rnd+`?vde=true`, nil, &parValue)
	if err != nil {
		t.Error(err)
		return
	}
	if parValue.Name != rnd {
		t.Error(fmt.Errorf(`wrong value of parameter`))
	}
	var tblResult tablesResult
	err = sendGet(`tables?vde=true`, nil, &tblResult)
	if err != nil {
		t.Error(err)
		return
	}
	if tblResult.Count < 5 {
		t.Error(fmt.Errorf(`wrong tables result`))
	}
	form = url.Values{"Name": {rnd}, `vde`: {`1`}, "Columns": {`[{"name":"MyName","type":"varchar", "index": "1",
		"conditions":"true"},
	  {"name":"Amount", "type":"number","index": "0", "conditions":"true"},
	  {"name":"Active", "type":"character","index": "0", "conditions":"true"}]`},
		"Permissions": {`{"insert": "true", "update" : "true", "new_column": "true"}`}}
	err = postTx(`NewTable`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	var tResult tableResult
	err = sendGet(`table/`+rnd+`?vde=true`, nil, &tResult)
	if err != nil {
		t.Error(err)
		return
	}
	if tResult.Name != rnd {
		t.Error(fmt.Errorf(`wrong table result`))
		return
	}
	var retList listResult
	err = sendGet(`list/contracts?vde=true`, nil, &retList)
	if err != nil {
		t.Error(err)
		return
	}
	if converter.StrToInt64(retList.Count) < 7 {
		t.Error(fmt.Errorf(`The number of records %s < 7`, retList.Count))
		return
	}
	var retRow rowResult
	err = sendGet(`row/contracts/2?vde=true`, nil, &retRow)
	if err != nil {
		t.Error(err)
		return
	}
	if !strings.Contains(retRow.Value[`value`], `VDEFunctions`) {
		t.Error(`wrong row result`)
		return
	}
	var retCont contractsResult
	err = sendGet(`contracts?vde=true`, nil, &retCont)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{`Value`: {`contract ` + rnd + ` {
		data {
			Par string
		}
		action { Test("active",  $Par)}}`}, `Conditions`: {`ContractConditions("MainCondition")`}, `vde`: {`true`}}

	if _, _, err = postTxResult(`NewContract`, &form); err != nil {
		t.Error(err)
		return
	}
	var cont getContractResult
	err = sendGet(`contract/`+rnd+`?vde=true`, nil, &cont)
	if err != nil {
		t.Error(err)
		return
	}
	if !strings.HasSuffix(cont.Name, rnd) {
		t.Error(`wrong contract result`)
		return
	}

	form = url.Values{"Name": {rnd}, "Value": {`Page`}, "Menu": {`government`},
		"Conditions": {`true`}, `vde`: {`1`}}
	err = postTx(`NewPage`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	err = sendPost(`content/page/`+rnd, &url.Values{`vde`: {`true`}}, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"Name": {rnd}, "Value": {`Menu`}, "Conditions": {`true`}, `vde`: {`1`}}
	err = postTx(`NewMenu`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	err = sendPost(`content/menu/`+rnd, &url.Values{`vde`: {`true`}}, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	//	t.Error(`OK`)
}
