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

	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/stretchr/testify/assert"
)

func TestNewEcosystem(t *testing.T) {
	var (
		err    error
		result string
	)
	if err = keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	form := url.Values{`Name`: {`test`}}
	if _, result, err = postTxResult(`NewEcosystem`, &form); err != nil {
		t.Error(err)
		return
	}
	var ret ecosystemsResult
	err = sendGet(`ecosystems`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if int64(ret.Number) != converter.StrToInt64(result) {
		t.Error(fmt.Errorf(`Ecosystems %d != %s`, ret.Number, result))
		return
	}

	form = url.Values{`Name`: {crypto.RandSeq(13)}}
	if err := postTx(`NewEcosystem`, &form); err != nil {
		t.Error(err)
		return
	}
}

func TestEditEcosystem(t *testing.T) {
	var (
		err error
	)
	if err = keyLogin(2); err != nil {
		t.Error(err)
		return
	}
	menu := `government`
	value := `P(test,test paragraph)`

	name := randName(`page`)
	form := url.Values{"Name": {name}, "Value": {value},
		"Menu": {menu}, "Conditions": {"ContractConditions(`MainCondition`)"}}
	err = postTx(`@1NewPage`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	err = postTx(`@1NewPage`, &form)
	if cutErr(err) != fmt.Sprintf(`{"type":"warning","error":"Page %s already exists"}`, name) {
		t.Error(err)
		return
	}
	form = url.Values{"Id": {`1`}, "Value": {value},
		"Menu": {menu}, "Conditions": {"ContractConditions(`MainCondition`)"}}
	err = postTx(`@1EditPage`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	name = randName(`test`)
	form = url.Values{"Value": {`contract ` + name + ` {
		action { Test("empty",  "empty value")}}`},
		"Conditions": {`ContractConditions("MainCondition")`}}
	_, id, err := postTxResult(`@1NewContract`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"Id": {id}, "Value": {`contract ` + name + ` {
		action { Test("empty3",  "empty value")}}`},
		"Conditions": {`ContractConditions("MainCondition")`}}
	if err := postTx(`@1EditContract`, &form); err != nil {
		t.Error(err)
		return
	}
}

func TestEcosystemParams(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	var ret ecosystemParamsResult
	err := sendGet(`ecosystemparams`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if len(ret.List) < 5 {
		t.Error(fmt.Errorf(`wrong count of parameters %d`, len(ret.List)))
	}
	err = sendGet(`ecosystemparams?names=ecosystem_name,new_table&ecosystem=1`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if len(ret.List) != 1 {
		t.Error(fmt.Errorf(`wrong count of parameters %d`, len(ret.List)))
	}

}

func TestSystemParams(t *testing.T) {

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}

	var ret ecosystemParamsResult

	err := sendGet(`systemparams`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, 62, len(ret.List), `wrong count of parameters %d`, len(ret.List))
}

func TestSomeSystemParam(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}

	var ret ecosystemParamsResult

	param := "gap_between_blocks"
	err := sendGet(`systemparams/?names=`+param, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, 1, len(ret.List), "parameter %s not found", param)
}

func TestEcosystemParam(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	var ret, ret1 paramValue
	err := sendGet(`ecosystemparam/changing_menu`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if ret.Value != `ContractConditions("MainCondition")` {
		t.Error(err)
		return
	}
	err = sendGet(`ecosystemparam/myval`, nil, &ret1)
	if err != nil && err.Error() != `400 {"error": "", "msg": "" }` {
		t.Error(err)
		return
	}
	if len(ret1.Value) != 0 {
		t.Error(err)
		return
	}
}

func TestAppParams(t *testing.T) {
	assert.NoError(t, keyLogin(1))

	rnd := `rnd` + crypto.RandSeq(3)
	form := url.Values{`App`: {`1`}, `Name`: {rnd + `1`}, `Value`: {`simple string,index`}, `Conditions`: {`true`}}
	assert.NoError(t, postTx(`NewAppParam`, &form))

	form[`Name`] = []string{rnd + `2`}
	form[`Value`] = []string{`another string`}
	assert.NoError(t, postTx(`NewAppParam`, &form))

	var ret appParamsResult
	assert.NoError(t, sendGet(`appparams/1`, nil, &ret))
	if len(ret.List) < 2 {
		t.Error(fmt.Errorf(`wrong count of parameters %d`, len(ret.List)))
	}

	assert.NoError(t, sendGet(fmt.Sprintf(`appparams/1?names=%s1,%[1]s2&ecosystem=1`, rnd), nil, &ret))
	assert.Len(t, ret.List, 2)

	var ret1, ret2 paramValue
	assert.NoError(t, sendGet(`appparam/1/`+rnd+`2`, nil, &ret1))
	assert.Equal(t, `another string`, ret1.Value)

	form[`Id`] = []string{ret1.ID}
	form[`Name`] = []string{rnd + `2`}
	form[`Value`] = []string{`{"par1":"value 1", "par2":"value 2"}`}
	assert.NoError(t, postTx(`EditAppParam`, &form))

	form = url.Values{"Value": {`contract ` + rnd + `Par { data {} conditions {} action
	{ var row map
		row=JSONDecode(AppParam(1, "` + rnd + `2"))
	    $result = row["par1"] }
	}`}, "Conditions": {"true"}}
	assert.NoError(t, postTx(`NewContract`, &form))

	_, msg, err := postTxResult(rnd+`Par`, &form)
	assert.NoError(t, err)
	assert.Equal(t, "value 1", msg)

	forTest := tplList{{`AppParam(` + rnd + `1, 1, Source: myname)`,
		`[{"tag":"data","attr":{"columns":["id","name"],"data":[["1","simple string"],["2","index"]],"source":"myname","types":["text","text"]}}]`},
		{`SetVar(myapp, 1)AppParam(` + rnd + `2, App: #myapp#)`,
			`[{"tag":"text","text":"{"par1":"value 1", "par2":"value 2"}"}]`}}
	for _, item := range forTest {
		var ret contentResult
		assert.NoError(t, sendPost(`content`, &url.Values{`template`: {item.input}}, &ret))
		assert.Equal(t, item.want, RawToString(ret.Tree))
	}

	assert.EqualError(t, sendGet(`appparam/1/myval`, nil, &ret2), `400 {"error": "", "msg": "" }`)
	assert.Len(t, ret2.Value, 0)
}
