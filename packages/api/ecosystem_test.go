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
	"github.com/stretchr/testify/require"
)

func TestNewEcosystem(t *testing.T) {
	var (
		err    error
		result string
	)

	require.NoError(t, keyLogin(1))

	form := url.Values{`Name`: {`test`}}
	_, result, err = postTxResult(`NewEcosystem`, &form)
	require.NoError(t, err)

	var ret ecosystemsResult
	require.NoError(t, sendGet(`ecosystems`, nil, &ret))

	require.Equalf(t, converter.StrToInt64(result), int64(ret.Number), `Ecosystems %d != %s`, ret.Number, result)

	form = url.Values{`Name`: {crypto.RandSeq(13)}}
	require.NoError(t, postTx(`NewEcosystem`, &form))
}

func TestEditEcosystem(t *testing.T) {
	var (
		err error
	)

	require.NoError(t, keyLogin(2))
	menu := `government`
	value := `P(test,test paragraph)`

	name := randName(`page`)
	form := url.Values{
		"Name":       {name},
		"Value":      {value},
		"Menu":       {menu},
		"Conditions": {"ContractConditions(`MainCondition`)"},
	}

	require.NoError(t, postTx(`@1NewPage`, &form))

	require.Equal(t, cutErr(postTx(`@1NewPage`, &form)), fmt.Sprintf(`{"type":"warning","error":"Page %s already exists"}`, name))

	form = url.Values{
		"Id":         {`1`},
		"Value":      {value},
		"Menu":       {menu},
		"Conditions": {"ContractConditions(`MainCondition`)"},
	}
	require.NoError(t, postTx(`@1EditPage`, &form))

	name = randName(`test`)
	form = url.Values{
		"Value":      {`contract ` + name + ` {action { Test("empty",  "empty value")}}`},
		"Conditions": {`ContractConditions("MainCondition")`},
	}

	_, id, err := postTxResult(`@1NewContract`, &form)
	require.NoError(t, err)

	form = url.Values{
		"Id":         {id},
		"Value":      {`contract ` + name + ` {action { Test("empty3",  "empty value")}}`},
		"Conditions": {`ContractConditions("MainCondition")`},
	}

	require.NoError(t, postTx(`@1EditContract`, &form))
}

func TestEcosystemParams(t *testing.T) {
	require.NoError(t, keyLogin(1))

	var ret ecosystemParamsResult
	require.NoError(t, sendGet(`ecosystemparams`, nil, &ret))

	assert.Truef(t, len(ret.List) >= 5, `wrong count of parameters %d`, len(ret.List))
	require.NoError(t, sendGet(`ecosystemparams?names=ecosystem_name,new_table&ecosystem=1`, nil, &ret))
	require.Equalf(t, len(ret.List), 1, `wrong count of parameters %d`, len(ret.List))
}

func TestSystemParams(t *testing.T) {

	require.NoError(t, keyLogin(1))

	var ret ecosystemParamsResult

	require.NoError(t, sendGet(`systemparams`, nil, &ret))
	assert.Equal(t, 62, len(ret.List), `wrong count of parameters %d`, len(ret.List))
}

func TestSomeSystemParam(t *testing.T) {
	require.NoError(t, keyLogin(1))

	var ret ecosystemParamsResult

	param := "gap_between_blocks"
	require.NoError(t, sendGet(`systemparams/?names=`+param, nil, &ret))
	assert.Equal(t, 1, len(ret.List), "parameter %s not found", param)
}

func TestEcosystemParam(t *testing.T) {
	require.NoError(t, keyLogin(1))

	var ret, ret1 paramValue
	require.NoError(t, sendGet(`ecosystemparam/changing_menu`, nil, &ret))
	require.Equal(t, `ContractConditions("MainCondition")`, ret.Value)

	err := sendGet(`ecosystemparam/myval`, nil, &ret1)
	require.Error(t, err, "must be error")
	require.EqualError(t, err, `400 {"error": "", "msg": "" }`)

	require.Equal(t, 0, len(ret1.Value), "returned value must be equal 0")
}

func TestAppParams(t *testing.T) {
	require.NoError(t, keyLogin(1))

	rnd := `rnd` + crypto.RandSeq(3)
	form := url.Values{`App`: {`1`}, `Name`: {rnd + `1`}, `Value`: {`simple string,index`}, `Conditions`: {`true`}}
	require.NoError(t, postTx(`NewAppParam`, &form))

	form[`Name`] = []string{rnd + `2`}
	form[`Value`] = []string{`another string`}
	require.NoError(t, postTx(`NewAppParam`, &form))

	var ret appParamsResult
	require.NoError(t, sendGet(`appparams/1`, nil, &ret))
	if len(ret.List) < 2 {
		t.Error(fmt.Errorf(`wrong count of parameters %d`, len(ret.List)))
	}

	require.NoError(t, sendGet(fmt.Sprintf(`appparams/1?names=%s1,%[1]s2&ecosystem=1`, rnd), nil, &ret))
	require.Len(t, ret.List, 2)

	var ret1, ret2 paramValue
	require.NoError(t, sendGet(`appparam/1/`+rnd+`2`, nil, &ret1))
	require.Equal(t, `another string`, ret1.Value)

	form[`Id`] = []string{ret1.ID}
	form[`Name`] = []string{rnd + `2`}
	form[`Value`] = []string{`{"par1":"value 1", "par2":"value 2"}`}
	assert.NoError(t, postTx(`EditAppParam`, &form))

	form = url.Values{"Value": {`contract ` + rnd + `Par { data {} conditions {} action
	{ var row map
		row=JSONDecode(AppParam(1, "` + rnd + `2"))
	    $result = row["par1"] }
	}`}, "Conditions": {"true"}}
	require.NoError(t, postTx(`NewContract`, &form))

	_, msg, err := postTxResult(rnd+`Par`, &form)
	require.NoError(t, err)
	require.Equal(t, "value 1", msg)

	forTest := tplList{{`AppParam(` + rnd + `1, 1, Source: myname)`,
		`[{"tag":"data","attr":{"columns":["id","name"],"data":[["1","simple string"],["2","index"]],"source":"myname","types":["text","text"]}}]`},
		{`AppParam(` + rnd + `2, App: 1)`,
			`[{"tag":"text","text":"{"par1":"value 1", "par2":"value 2"}"}]`}}
	for _, item := range forTest {
		var ret contentResult
		require.NoError(t, sendPost(`content`, &url.Values{`template`: {item.input}}, &ret))
		require.Equal(t, item.want, RawToString(ret.Tree))
	}

	require.EqualError(t, sendGet(`appparam/1/myval`, nil, &ret2), `400 {"error": "", "msg": "" }`)
	require.Len(t, ret2.Value, 0)
}
