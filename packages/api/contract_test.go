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

	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdate_FullNodes(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}

	err := postTx("UpdateSysParam", &url.Values{
		"Name":  {"full_nodes"},
		"Value": {"[]"},
	})
	if err != nil {
		t.Error(err)
		return
	}
}
func TestHardContract(t *testing.T) {
	assert.NoError(t, keyLogin(1))

	rnd := `hard` + crypto.RandSeq(4)
	form := url.Values{`Value`: {`contract ` + rnd + ` {
		    data {
			}
			action { 
				var i int
				while i < 200 {
				 DBFind("pages").Where("id=5")
				 DBUpdate("pages", 5, "value", "P(text)")
				 DBInsert("pages", "name,value,conditions", Sprintf("` + rnd + `%d", i), "P(text)","true")
				 DBFind("pages").Where("id=6")
				 DBUpdate("pages", 6, "value", "P(text)")
				 i = i + 1
			   }
			}}`}, "ApplicationId": {"1"}, `Conditions`: {`true`}}
	assert.NoError(t, postTx(`NewContract`, &form))
	assert.EqualError(t, postTx(rnd, &url.Values{}), `{"type":"txError","error":"Time limit exceeded"}`)
}

func TestExistContract(t *testing.T) {
	assert.NoError(t, keyLogin(1))
	form := url.Values{"Name": {`EditPage`}, "Value": {`contract EditPage {action {}}`},
		"ApplicationId": {`1`}, "Conditions": {`true`}}
	err := postTx(`NewContract`, &form)
	assert.EqualError(t, err, `{"type":"panic","error":"Contract EditPage already exists"}`)
}

func TestNewContracts(t *testing.T) {

	wanted := func(name, want string) bool {
		var ret getTestResult
		return assert.NoError(t, sendPost(`test/`+name, nil, &ret)) && assert.Equal(t, want, ret.Value)
	}

	assert.NoError(t, keyLogin(1))
	rnd := crypto.RandSeq(4)
	for i, item := range contracts {
		var ret getContractResult
		if i > 100 {
			break
		}
		name := strings.Replace(item.Name, `#rnd#`, rnd, -1)
		err := sendGet(`contract/`+name, nil, &ret)
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(apiErrors[`E_CONTRACT`], name)) {
				form := url.Values{"Name": {name}, "Value": {strings.Replace(item.Value,
					`#rnd#`, rnd, -1)},
					"ApplicationId": {`1`}, "Conditions": {`true`}}
				if err := postTx(`NewContract`, &form); err != nil {
					assert.EqualError(t, err, item.Params[0].Results[`error`])
					continue
				}
			} else {
				t.Error(err)
				return
			}
		}
		if strings.HasSuffix(name, `testUpd`) {
			continue
		}
		for _, par := range item.Params {
			form := url.Values{}
			for key, value := range par.Params {
				form[key] = []string{value}
			}
			if err := postTx(name, &form); err != nil {
				assert.EqualError(t, err, par.Results[`error`])
				continue
			}
			for key, value := range par.Results {
				if !wanted(key, value) {
					return
				}
			}
		}
	}
	var row rowResult
	assert.NoError(t, sendGet(`row/menu/1`, nil, &row))
	assert.NotEqual(t, `update`, row.Value[`value`])
}

var contracts = []smartContract{
	{`StrNil`, `contract StrNil {
		action {
			Test("result", Sprintf("empty: %s", Str(nil)))
		}
	}`, []smartParams{
		{nil, map[string]string{`result`: `empty: `}},
	}},
	{`TestJSON`, `contract TestJSON {
		data {}
		conditions { }
		action {
		   var a map
		   a["ok"] = 10
		   a["arr"] = ["first", "<second>"]
		   Test("json", JSONEncode(a))
		   Test("ok", JSONEncodeIndent(a, "\t"))
		}
	}`, []smartParams{
		{nil, map[string]string{`ok`: "{\n\t\"arr\": [\n\t\t\"first\",\n\t\t\"<second>\"\n\t],\n\t\"ok\": 10\n}",
			`json`: "{\"arr\":[\"first\",\"<second>\"],\"ok\":10}"}},
	}},
	{`GuestKey`, `contract GuestKey {
		action {
			Test("result", $guest_key)
		}
	}`, []smartParams{
		{nil, map[string]string{`result`: `4544233900443112470`}},
	}},
	{`TestCyr`, `contract TestCyr {
		data {}
		conditions { }
		action {
		   //тест
		   var a map
		   a["тест"] = "тест"
		   Test("ok", a["тест"])
		}
	}`, []smartParams{
		{nil, map[string]string{`ok`: `тест`}},
	}},
	{`DBFindLike`, `contract DBFindLike {
		action {
			var list array
			list = DBFind("pages").Where({"name":{"$like": "ort_"}})
			Test("size", Len(list))
			list = DBFind("pages").Where({"name":{"$end": "page"}})
			Test("end", Len(list))
		}
	}`, []smartParams{
		{nil, map[string]string{`size`: `2`, `end`: `1`}},
	}},
	{`TestDBFindOK`, `
			contract TestDBFindOK {
			action {
				var ret array
				var vals map
				ret = DBFind("contracts").Columns("id,value").Where({"$and":[{"id":{"$gte": 3}}, {"id":{"$lte":5}}]}).Order("id")
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
				ret = DBFind("contracts").Columns("id").Order(["id"]).Offset(1).Limit(1)
				if Len(ret) != 1 {
					Test("3",  "0")
				} else {
					vals = ret[0]
					Test("3", vals["id"])
				}
				ret = DBFind("contracts").Columns("id").Where({"$or":[{"id": "1"}]})
				if Len(ret) != 1 {
					Test("4",  "0")
				} else {
					vals = ret[0]
					Test("4", vals["id"])
				}
				ret = DBFind("contracts").Columns("id").Where({"id": 1})
				if Len(ret) != 1 {
					Test("4",  "0")
				} else {
					vals = ret[0]
					Test("4", vals["id"])
				}
				ret = DBFind("contracts").Columns("id,value").Where({"id":[{"$gt":3},{"$lt":8}]}).Order([{"id": 1}, {"name": "-1"}])
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
				var one string
				one = DBFind("contracts").WhereId(5).One("id")
				Test("7",  one)
				var row map
				row = DBFind("contracts").WhereId(3).Row()
				Test("8",  row["id"])
				Test("255",  "255")
			}
		}`,
		[]smartParams{
			{nil, map[string]string{`0`: `1`, `1`: `1`, `2`: `2`, `3`: `2`, `4`: `1`, `5`: `4`,
				`6`: `7`, `7`: `5`, `8`: `3`, `255`: `255`}},
		}},
	{`DBFindCol`, `contract DBFindCol {
		action {
			var ret string
			ret = DBFind("keys").Columns(["amount", "id"]).One("amount")
			Test("size", Size(ret)>0)
		}
	}`, []smartParams{
		{nil, map[string]string{`size`: `true`}},
	}},
	{`DBFindColumnNow`, `contract DBFindColumnNow {
		action {
			var list array
			list = DBFind("keys").Columns("now()")
		}
	}`, []smartParams{
		{nil, map[string]string{`error`: `{"type":"panic","error":"pq: current transaction is aborted, commands ignored until end of transaction block"}`}},
	}},
	{`DBFindCURRENT`, `contract DBFindCURRENT {
		action {
			var list array
			list = DBFind("mytable").Where({"date": {"$lt": "CURRENT_DATE"}})
		}
	}`, []smartParams{
		{nil, map[string]string{`error`: `{"type":"panic","error":"It is prohibited to use NOW() or current time functions"}`}},
	}},
	{`RowType`, `contract RowType {
	action {
		var app map
		var result string
		result = GetType(app)
		app = DBFind("applications").Where({"id":"1"}).Row()
		result = result + GetType(app)
		app["app_id"] = 2
		Test("result", Sprintf("%s %s %d", result, app["name"], app["app_id"]))
	}
}`, []smartParams{
		{nil, map[string]string{`result`: `map[string]interface {}map[string]interface {} System 2`}},
	}},
	{`StackType`, `contract StackType {
		action {
			var lenStack int
			lenStack = Len($stack)
			var par string
			par = $stack[0]
			Test("result", Sprintf("len=%d %v %s", lenStack, $stack, par))
		}
	}`, []smartParams{
		{nil, map[string]string{`result`: `len=1 [@1StackType] @1StackType`}},
	}},
	{`DBFindNow`, `contract DBFindNow {
		action {
			var list array
			list = DBFind("mytable").Where({"date": {"$lt": "now ( )"}})
		}
	}`, []smartParams{
		{nil, map[string]string{`error`: `{"type":"panic","error":"It is prohibited to use NOW() or current time functions"}`}},
	}},
	{`BlockTimeCheck`, `contract BlockTimeCheck {
		action {
			if Size(BlockTime()) == Size("2006-01-02 15:04:05") {
				Test("ok", "1")
			} else {
				Test("ok", "0")
			}
		}
	}`, []smartParams{
		{nil, map[string]string{`ok`: `1`}},
	}},
	{`RecCall`, `contract RecCall {
		data {    }
		conditions {    }
		action {
			var par map
			CallContract("RecCall", par)
		}
	}`, []smartParams{
		{nil, map[string]string{`error`: `{"type":"panic","error":"There is loop in @1RecCall contract"}`}},
	}},
	{`Recursion`, `contract Recursion {
		data {    }
		conditions {    }
		action {
			Recursion()
		}
	}`, []smartParams{
		{nil, map[string]string{`error`: `{"type":"panic","error":"The contract can't call itself recursively"}`}},
	}},
	{`MyTable#rnd#`, `contract MyTable#rnd# {
		action {
			NewTable("Name,Columns,ApplicationId,Permissions", "#rnd#1", 
				"[{\"name\":\"MyName\",\"type\":\"varchar\", \"index\": \"0\", \"conditions\":{\"update\":\"true\", \"read\":\"true\"}}]", 100,
				 "{\"insert\": \"true\", \"update\" : \"true\", \"new_column\": \"true\"}")
			var cols array
			cols[0] = "{\"conditions\":\"true\",\"name\":\"column1\",\"type\":\"text\"}"
			cols[1] = "{\"conditions\":\"true\",\"name\":\"column2\",\"type\":\"text\"}"
			NewTable("Name,Columns,ApplicationId,Permissions", "#rnd#2", 
				JSONEncode(cols), 100,
				 "{\"insert\": \"true\", \"update\" : \"true\", \"new_column\": \"true\"}")
			
			Test("ok", "1")
		}
	}`, []smartParams{
		{nil, map[string]string{`ok`: `1`}},
	}},
	{`IntOver`, `contract IntOver {
				action {
					info Int("123456789101112131415161718192021222324252627282930")
				}
			}`, []smartParams{
		{nil, map[string]string{`error`: `{"type":"panic","error":"123456789101112131415161718192021222324252627282930 is not a valid integer : value out of range"}`}},
	}},
	{`Double`, `contract Double {
		data {    }
		conditions {    }
		action {
			$$$$$$$$result = "hello"
		}
	}`, []smartParams{
		{nil, map[string]string{`error`: `{"type":"panic","error":"unknown lexem $ [Ln:5 Col:6]"}`}},
	}},
	{`Price`, `contract Price {
		action {
			Test("int", Int("")+Int(nil)+2)
			Test("price", 1)
		}
		func price() money {
			return Money(100)
		}
	}`, []smartParams{
		{nil, map[string]string{`price`: `1`, `int`: `2`}},
	}},
	{`CheckFloat`, `contract CheckFloat {
			action {
			var fl float
			fl = -3.67
			Test("float2", Sprintf("%d %s", Int(1.2), Str(fl)))
			Test("float3", Sprintf("%.2f %.2f", 10.7/7, 10/7.0))
			Test("float4", Sprintf("%.2f %.2f %.2f", 10+7.0, 10-3.1, 5*2.5))
			Test("float5", Sprintf("%t %t %t %t %t", 10 <= 7.0, 4.5 <= 5, 3>5.7, 6 == 6.0, 7 != 7.1))
		}}`, []smartParams{
		{nil, map[string]string{`float2`: `1 -3.670000`, `float3`: `1.53 1.43`, `float4`: `17.00 6.90 12.50`, `float5`: `false true false true true`}},
	}},
	{`Crash`, `contract Crash { data {} conditions {} action

			{ $result=DBUpdate("menu", 1, {"value": "updated"}) }
			}`,
		[]smartParams{
			{nil, map[string]string{`error`: `{"type":"panic","error":"runtime panic error"}`}},
		}},
	{`TestOneInput`, `contract TestOneInput {
			data {
				list array
			}
			action {
				var coltype string
				coltype = GetColumnType("keys", "amount" )
				Test("oneinput",  $list[0]+coltype)
			}
		}`,
		[]smartParams{
			{map[string]string{`list`: `Input value`}, map[string]string{`oneinput`: `Input valuemoney`}},
		}},
	{`DBProblem`, `contract DBProblem {
		action{
			DBFind("members1").Where({"member_name": "name"})
		}
	}`,
		[]smartParams{
			{nil, map[string]string{`error`: `{"type":"panic","error":"pq: current transaction is aborted, commands ignored until end of transaction block"}`}},
		}},
	{`TestMultiForm`, `contract TestMultiForm {
				data {
					list array
				}
				action {
					Test("multiform",  $list[0]+$list[1])
				}
			}`,
		[]smartParams{
			{map[string]string{`list[]`: `2`, `list[0]`: `start`, `list[1]`: `finish`}, map[string]string{`multiform`: `startfinish`}},
		}},
	{`errTestMessage`, `contract errTestMessage {
			conditions {
			}
			action { qvar ivar int}
		}`,
		[]smartParams{
			{nil, map[string]string{`error`: `{"type":"panic","error":"unknown variable qvar"}`}},
		}},

	{`EditProfile9`, `contract EditProfile9 {
			data {
			}
			conditions {
			}
			action {
				var ar array
				ar = Split("point 1,point 2", ",")
				Test("split",  Str(ar[1]))
				$ret = DBFind("contracts").Columns("id,value").Where({"id":[{"$gte": 3}, {"$lte":5}]}).Order("id")
				Test("edit",  "edit value 0")
			}
		}`,
		[]smartParams{
			{nil, map[string]string{`edit`: `edit value 0`, `split`: `point 2`}},
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
	{`testLong`, `contract testLong {
			action { Test("long",  "long result")
				$result = DBFind("contracts").WhereId(2).One("value") + DBFind("contracts").WhereId(4).One("value")
				Println("Result", $result)
				Test("long",  "long result")
			}}`,
		[]smartParams{
			{nil, map[string]string{`long`: `long result`}},
		}},
	{`testSimple`, `contract testSimple {
					data {
						Amount int
						Name   string
					}
					conditions {
						Test("scond", $Amount, $Name)
					}
					action { Test("sact", $Name, $Amount)}}`,
		[]smartParams{
			{map[string]string{`Name`: `Simple name`, `Amount`: `-56781`},
				map[string]string{`scond`: `-56781Simple name`,
					`sact`: `Simple name-56781`}},
		}},
	{`errTestVar`, `contract errTestVar {
				conditions {
				}
				action { var ivar int}
			}`,
		nil},
	{`testGetContract`, `contract testGetContract {
			action { Test("ByName", GetContractByName(""), GetContractByName("ActivateContract"))
				Test("ById", GetContractById(10000000), GetContractById(16))}}`,
		[]smartParams{
			{nil, map[string]string{`ByName`: `0 2`,
				`ById`: `EditLang`}},
		}},
	{
		`testDateTime`, `contract testDateTime {
				data {
					Date string
					Unix int
				}
				action {
					Test("DateTime", DateTime($Unix))
					Test("UnixDateTime", UnixDateTime($Date))
				}
			}`,
		[]smartParams{
			{map[string]string{
				"Unix": "1257894000",
				"Date": "2009-11-11 04:00:00",
			}, map[string]string{
				"DateTime":     "2009-11-11 04:00:00",
				"UnixDateTime": timeMustParse("2009-11-11 04:00:00"),
			}},
		},
	},
}

func timeMustParse(value string) string {
	t, _ := time.Parse("2006-01-02 15:04:05", value)
	return converter.Int64ToStr(t.Unix())
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
	form := url.Values{`Id`: {sid}, `Value`: {newCode}, `Conditions`: {row.Value[`conditions`]}, `WalletId`: {"01231234123412341230"}}
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

func TestNewTableWithEmptyName(t *testing.T) {
	require.NoError(t, keyLogin(1))
	sql1 := `new_column varchar(10); update block_chain set key_id='1234' where id='1' --`
	sql2 := `new_column varchar(10); update block_chain set key_id='12' where id='1' --`
	name := randName(`tbl`)
	form := url.Values{
		"Name":          {name},
		"Columns":       {"[{\"name\":\"" + sql1 + "\",\"type\":\"varchar\", \"index\": \"0\", \"conditions\":{\"update\":\"true\", \"read\":\"true\"}}]"},
		"ApplicationId": {"1"},
		"Permissions":   {"{\"insert\": \"true\", \"update\" : \"true\", \"new_column\": \"true\"}"},
	}

	require.NoError(t, postTx("NewTable", &form))

	form = url.Values{"TableName": {name}, "Name": {sql2},
		"Type": {"varchar"}, "Index": {"0"}, "Permissions": {"true"}}
	assert.NoError(t, postTx(`NewColumn`, &form))

	form = url.Values{
		"Name":          {""},
		"Columns":       {"[{\"name\":\"MyName\",\"type\":\"varchar\", \"index\": \"0\", \"conditions\":{\"update\":\"true\", \"read\":\"true\"}}]"},
		"ApplicationId": {"1"},
		"Permissions":   {"{\"insert\": \"true\", \"update\" : \"true\", \"new_column\": \"true\"}"},
	}

	if err := postTx("NewTable", &form); err == nil || err.Error() !=
		`400 {"error": "E_SERVER", "msg": "Name is empty" }` {
		t.Error(`wrong error`, err)
	}

	form = url.Values{
		"Name":          {"Digit" + name},
		"Columns":       {"[{\"name\":\"1\",\"type\":\"varchar\", \"index\": \"0\", \"conditions\":{\"update\":\"true\", \"read\":\"true\"}}]"},
		"ApplicationId": {"1"},
		"Permissions":   {"{\"insert\": \"true\", \"update\" : \"true\", \"new_column\": \"true\"}"},
	}

	assert.EqualError(t, postTx("NewTable", &form), `{"type":"panic","error":"Column name cannot begin with digit"}`)
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
			action { Test("active",  $Par)}}`}, "ApplicationId": {"1"}, `Conditions`: {`true`}}
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

func TestDeactivateContracts(t *testing.T) {

	wanted := func(name, want string) bool {
		var ret getTestResult
		return assert.NoError(t, sendPost(`test/`+name, nil, &ret)) && assert.Equal(t, want, ret.Value)
	}

	assert.NoError(t, keyLogin(1))

	rnd := `rnd` + crypto.RandSeq(6)
	form := url.Values{`Value`: {`contract ` + rnd + ` {
		    data {
				Par string
			}
			action { Test("active",  $Par)}}`}, "ApplicationId": {"1"}, `Conditions`: {`true`}}
	assert.NoError(t, postTx(`NewContract`, &form))

	var ret getContractResult
	assert.NoError(t, sendGet(`contract/`+rnd, nil, &ret))

	assert.NoError(t, postTx(`ActivateContract`, &url.Values{`Id`: {ret.TableID}}))
	assert.NoError(t, sendGet(`contract/`+rnd, nil, &ret))
	assert.True(t, ret.Active, `Not activate `+rnd)

	var row rowResult
	assert.NoError(t, sendGet(`row/contracts/`+ret.TableID, nil, &row))
	assert.Equal(t, "1", row.Value[`active`], `row not activate `+rnd)

	assert.NoError(t, postTx(rnd, &url.Values{`Par`: {rnd}}))

	if !wanted(`active`, rnd) {
		return
	}

	assert.NoError(t, postTx(`DeactivateContract`, &url.Values{`Id`: {ret.TableID}}))

	assert.NoError(t, sendGet(`contract/`+rnd, nil, &ret))
	assert.False(t, ret.Active, `Not deactivate `+rnd)

	var row2 rowResult
	assert.NoError(t, sendGet(`row/contracts/`+ret.TableID, nil, &row2))
	assert.Equal(t, "0", row2.Value[`active`])
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
			}}`}, "ApplicationId": {"1"}, `Conditions`: {`true`}}
	if err := postTx(`NewContract`, &form); err != nil {
		t.Error(err)
		return
	}
	form = url.Values{`Value`: {`contract ` + rnd + `Test {
			data {
				Recipient int "hidden"
				Amount  money
				Signature string "signature:` + rnd + `Transfer"
			}
			func action {
				` + rnd + `Transfer("Recipient,Amount,Signature",$Recipient,$Amount,$Signature )
				$result = "OOOPS " + Str($Amount)
			}
		  }
		`}, `Conditions`: {`true`}, "ApplicationId": {"1"}}
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
	imp = `{
		"menus": [
			{
				"Name": "test_%s",
				"Conditions": "ContractAccess(\"@1EditMenu\")",
				"Value": "MenuItem(main, Default Ecosystem Menu)"
			}
		],
		"contracts": [
			{
				"Name": "testContract%[1]s",
				"Value": "contract testContract%[1]s {\n    data {}\n    conditions {}\n    action {\n        var res array\n        res = DBFind(\"pages\").Columns(\"name\").Where({id: 1}).Order(\"id\")\n        $result = res\n    }\n    }",
				"Conditions": "ContractConditions(` + "`MainCondition`" + `)"
			}
		],
		"pages": [
			{
				"Name": "test_%[1]s",
				"Conditions": "ContractAccess(\"@1EditPage\")",
				"Menu": "default_menu",
				"Value": "P(class, Default Ecosystem Page)\nImage().Style(width:100px;)"
			}
		],
		"blocks": [
			{
				"Name": "test_%[1]s",
				"Conditions": "true",
				"Value": "block content"
			},
			{
				"Name": "test_a%[1]s",
				"Conditions": "true",
				"Value": "block content"
			},
			{
				"Name": "test_b%[1]s",
				"Conditions": "true",
				"Value": "block content"
			}
		],
		"tables": [
			{
				"Name": "members%[1]s",
				"Columns": "[{\"name\":\"name\",\"type\":\"varchar\",\"conditions\":\"true\"},{\"name\":\"birthday\",\"type\":\"datetime\",\"conditions\":\"true\"},{\"name\":\"member_id\",\"type\":\"number\",\"conditions\":\"true\"},{\"name\":\"val\",\"type\":\"text\",\"conditions\":\"true\"},{\"name\":\"name_first\",\"type\":\"text\",\"conditions\":\"true\"},{\"name\":\"name_middle\",\"type\":\"text\",\"conditions\":\"true\"}]",
				"Permissions": "{\"insert\":\"true\",\"update\":\"true\",\"new_column\":\"true\"}"
			}
		],
		"parameters": [
			{
				"Name": "host%[1]s",
				"Value": "",
				"Conditions": "ContractConditions(` + "`MainCondition`" + `)"
			},
			{
				"Name": "host0%[1]s",
				"Value": "Русский текст",
				"Conditions": "ContractConditions(` + "`MainCondition`" + `)"
			}
		],
		"data": [
			{
				"Table": "members%[1]s",
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
	name := crypto.RandSeq(4)
	form := url.Values{"Data": {fmt.Sprintf(imp, name)}}
	err := postTx(`@1Import`, &form)
	if err != nil {
		t.Error(err)
		return
	}

}

func TestEditContracts_ChangeWallet(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	rnd := `rnd` + crypto.RandSeq(6)
	code := `contract ` + rnd + ` {
		data {
			Par string "optional"
		}
		action { $result = $par}}`
	form := url.Values{`Value`: {code}, `Conditions`: {`true`}}
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
	keyID := ret.WalletID
	sid := ret.TableID
	var row rowResult
	err = sendGet(`row/contracts/`+sid, nil, &row)
	if err != nil {
		t.Error(err)
		return
	}

	if err := postTx(`ActivateContract`, &url.Values{`Id`: {sid}}); err != nil {
		t.Error(err)
		return
	}

	code = row.Value[`value`]
	form = url.Values{`Id`: {sid}, `Value`: {code}, `Conditions`: {row.Value[`conditions`]}, `WalletId`: {"1248-5499-7861-4204-5166"}}
	err = postTx(`EditContract`, &form)
	if err == nil {
		t.Error("Expected `Contract activated` error")
		return
	}
	err = sendGet(`contract/`+rnd, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if ret.WalletID != keyID {
		t.Error(`wrong walletID`, ret.WalletID, keyID)
		return
	}
	if err := postTx(`DeactivateContract`, &url.Values{`Id`: {sid}}); err != nil {
		t.Error(err)
		return
	}

	if err := postTx(`EditContract`, &form); err != nil {
		t.Error(err)
		return
	}
	err = sendGet(`contract/`+rnd, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if ret.Address != "1248-5499-7861-4204-5166" {
		t.Error(`wrong address`, ret.Address, "!= 1248-5499-7861-4204-5166")
		return
	}
}

func TestUpdateFunc(t *testing.T) {
	assert.NoError(t, keyLogin(1))

	rnd := `rnd` + crypto.RandSeq(6)
	form := url.Values{`Value`: {`contract f` + rnd + ` {
		data {
			par string
		}
		func action {
			$result = Sprintf("X=%s %s %s", $par, $original_contract, $this_contract)
		}}`}, "ApplicationId": {"1"}, `Conditions`: {`true`}}
	_, id, err := postTxResult(`NewContract`, &form)
	assert.NoError(t, err)

	form = url.Values{`Value`: {`
		contract one` + rnd + ` {
			action {
				var ret map
				ret = DBFind("contracts").Columns("id,value").WhereId(10).Row()
				$result = ret["id"]
		}}`}, "ApplicationId": {"1"}, `Conditions`: {`true`}}
	assert.NoError(t, postTx(`NewContract`, &form))

	form = url.Values{`Value`: {`contract row` + rnd + ` {
				action {
					var ret string
					ret = DBFind("contracts").Columns("id,value").WhereId(11).One("id")
					$result = ret
				}}
		
			`}, "ApplicationId": {"1"}, `Conditions`: {`true`}}
	assert.NoError(t, postTx(`NewContract`, &form))

	_, msg, err := postTxResult(`one`+rnd, &url.Values{})
	assert.NoError(t, err)
	assert.Equal(t, "10", msg)

	_, msg, err = postTxResult(`row`+rnd, &url.Values{})
	assert.NoError(t, err)
	assert.Equal(t, "11", msg)

	form = url.Values{`Value`: {`
		contract ` + rnd + ` {
		    data {
				Par string
			}
			action {
				$result = f` + rnd + `("par",$Par) + " " + $this_contract
			}}
		`}, "ApplicationId": {"1"}, `Conditions`: {`true`}}
	_, idcnt, err := postTxResult(`NewContract`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	_, msg, err = postTxResult(rnd, &url.Values{`Par`: {`my param`}})
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(`X=my param %s f%[1]s %[1]s`, rnd), msg)

	form = url.Values{`Id`: {id}, `Value`: {`
		func MyTest2(input string) string {
			return "Y="+input
		}`}, `Conditions`: {`true`}}
	err = postTx(`EditContract`, &form)
	assert.EqualError(t, postTx(`EditContract`, &form), `{"type":"panic","error":"Contracts or functions names cannot be changed"}`)

	form = url.Values{`Id`: {id}, `Value`: {`contract f` + rnd + `{
		data {
			par string
		}
		action {
			$result = "Y="+$par
		}}`}, `Conditions`: {`true`}}
	assert.NoError(t, postTx(`EditContract`, &form))

	_, msg, err = postTxResult(rnd, &url.Values{`Par`: {`new param`}})
	assert.NoError(t, err)
	assert.Equal(t, `Y=new param `+rnd, msg)

	form = url.Values{`Id`: {idcnt}, `Value`: {`
		contract ` + rnd + ` {
		    data {
				Par string
			}
			action {
				$result = f` + rnd + `("par",$Par) + f` + rnd + `("par","OK")
			}}
		`}, `Conditions`: {`true`}}
	_, idcnt, err = postTxResult(`EditContract`, &form)
	assert.NoError(t, err)

	_, msg, err = postTxResult(rnd, &url.Values{`Par`: {`finish`}})
	assert.NoError(t, err)
	assert.Equal(t, `Y=finishY=OK`, msg)
}

func TestGlobalVars(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	rnd := `rnd` + crypto.RandSeq(4)

	form := url.Values{`Value`: {`
		contract ` + rnd + ` {
		    data {
				Par string
			}
			action {
				$Par = $Par + "end"
				$key_id = 1234
				$result = Str($key_id) + $Par
			}}
		`}, "ApplicationId": {"1"}, `Conditions`: {`true`}}
	err := postTx(`NewContract`, &form)
	if err == nil {
		t.Errorf(`must be error`)
		return
	} else if err.Error() != `{"type":"panic","error":"system variable $key_id cannot be changed"}` {
		t.Error(err)
		return
	}
	form = url.Values{`Value`: {`contract c_` + rnd + ` {
		data { Test string }
		action {
			$result = $Test + Str($ecosystem_id)
		}
	}`}, "ApplicationId": {"1"}, `Conditions`: {`true`}}
	err = postTx(`NewContract`, &form)
	if err != nil {
		t.Error(err)
		return
	}

	form = url.Values{`Value`: {`
		contract a_` + rnd + ` {
			data { Par string}
			conditions {}
			action {
				var params map
				params["Test"] = "TEST"
				$aaa = 123
				if $Par == "b" {
				    $result = CallContract("b_` + rnd + `", params)
				} else {
				    $result = CallContract("c_` + rnd + `", params) + c_` + rnd + `("Test","OK")
				}
			}
		}`}, "ApplicationId": {"1"}, `Conditions`: {`true`}}
	err = postTx(`NewContract`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{`Value`: {`contract b_` + rnd + ` {
			data { Test string }
			action {
				$result = $Test + $aaa
			}
		}`}, "ApplicationId": {"1"}, `Conditions`: {`true`}}
	err = postTx(`NewContract`, &form)
	if err != nil {
		t.Error(err)
		return
	}

	err = postTx(`a_`+rnd, &url.Values{"Par": {"b"}})
	if err == nil {
		t.Errorf(`must be error aaa`)
		return
	} else if err.Error() != `{"type":"panic","error":"unknown extend identifier aaa"}` {
		t.Error(err)
		return
	}
	_, msg, err := postTxResult(`a_`+rnd, &url.Values{"Par": {"c"}})
	if err != nil {
		t.Error(err)
		return
	}
	if msg != `TEST1OK1` {
		t.Errorf(`wrong result %s`, msg)
		return
	}
}

func TestContractChain(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	rnd := `rnd` + crypto.RandSeq(4)

	form := url.Values{"Name": {rnd}, "ApplicationId": {"1"}, "Columns": {`[{"name":"value","type":"varchar", "index": "0", 
	  "conditions":"true"},
	{"name":"amount", "type":"number","index": "0", "conditions":"true"},
	{"name":"dt","type":"datetime", "index": "0", "conditions":"true"}]`},
		"Permissions": {`{"insert": "true", "update" : "true", "new_column": "true"}`}}
	err := postTx(`NewTable`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{`Value`: {`contract sub` + rnd + ` {
		data {
			Id int
		}
		action {
			$row = DBFind("` + rnd + `").Columns("value").WhereId($Id)
			if Len($row) != 1 {
				error "sub contract getting error"
			}
			$record = $row[0]
			$new = $record["value"]
			var val string
			val = $new+"="+$new
			DBUpdate("` + rnd + `", $Id, {"value": val })
		}
	}`}, "ApplicationId": {"1"}, `Conditions`: {`true`}}
	err = postTx(`NewContract`, &form)
	if err != nil {
		t.Error(err)
		return
	}

	form = url.Values{`Value`: {`contract ` + rnd + ` {
		data {
			Initial string
		}
		action {
			$id = DBInsert("` + rnd + `", {value: $Initial, amount:"0"})
			sub` + rnd + `("Id", $id)
			$row = DBFind("` + rnd + `").Columns("value").WhereId($id)
			if Len($row) != 1 {
				error "contract getting error"
			}
			$record = $row[0]
			$result = $record["value"]
		}
	}
		`}, "ApplicationId": {"1"}, `Conditions`: {`true`}}
	err = postTx(`NewContract`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	_, msg, err := postTxResult(rnd, &url.Values{`Initial`: {rnd}})
	if err != nil {
		t.Error(err)
		return
	}
	if msg != rnd+`=`+rnd {
		t.Error(fmt.Errorf(`wrong result %s`, msg))
	}

	form = url.Values{`Value`: {`contract ` + rnd + `1 {
		action {
			DBInsert("` + rnd + `", {amount: 0,dt: "timestamp NOW()"})
		}
	}
		`}, "ApplicationId": {"1"}, `Conditions`: {`true`}}
	assert.NoError(t, postTx(`NewContract`, &form))
	assert.EqualError(t, postTx(rnd+`1`, &url.Values{}),
		`{"type":"panic","error":"It is prohibited to use Now() function"}`)
}

func TestLoopCond(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	rnd := `rnd` + crypto.RandSeq(4)

	form := url.Values{`Value`: {`contract ` + rnd + `1 {
		conditions {
	    
		}
	}`}, `Conditions`: {`true`}, `ApplicationId`: {`1`}}
	err := postTx(`NewContract`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{`Value`: {`contract ` + rnd + `2 {
				conditions {
					ContractConditions("` + rnd + `1")
				}
			}`}, `Conditions`: {`true`}, `ApplicationId`: {`1`}}
	err = postTx(`NewContract`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	var ret getContractResult
	err = sendGet(`contract/`+rnd+`1`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	sid := ret.TableID
	form = url.Values{`Value`: {`contract ` + rnd + `1 {
				conditions {
					ContractConditions("` + rnd + `2")
				}
			}`}, `Id`: {sid}, `Conditions`: {`true`}, `ApplicationId`: {`1`}}
	err = postTx(`EditContract`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	assert.EqualError(t, postTx(rnd+`2`, &url.Values{}), `{"type":"panic","error":"There is loop in `+rnd+`1 contract"}`)

	form = url.Values{"Name": {`ecosystems`}, "InsertPerm": {`ContractConditions("MainCondition")`},
		"UpdatePerm":    {`EditEcosysName(1, "HANG")`},
		"NewColumnPerm": {`ContractConditions("MainCondition")`}}
	assert.NoError(t, postTx(`EditTable`, &form))
	assert.EqualError(t, postTx(`EditEcosystemName`, &url.Values{"EcosystemID": {`1`},
		"NewName": {`Hang`}}), `{"type":"panic","error":"There is loop in EditEcosysName contract"}`)

	form = url.Values{`Value`: {`contract ` + rnd + `shutdown {
		action
		{ DBInsert("` + rnd + `table", {"test": "SHUTDOWN"}) }
	}`}, `Conditions`: {`true`}, `ApplicationId`: {`1`}}
	assert.NoError(t, postTx(`NewContract`, &form))

	form = url.Values{
		"Name":          {rnd + `table`},
		"Columns":       {`[{"name":"test","type":"varchar", "index": "0", "conditions":"true"}]`},
		"ApplicationId": {"1"},
		"Permissions":   {`{"insert": "` + rnd + `shutdown()", "update" : "true", "new_column": "true"}`},
	}
	require.NoError(t, postTx("NewTable", &form))

	assert.EqualError(t, postTx(rnd+`shutdown`, &url.Values{}), `{"type":"panic","error":"There is loop in @1`+rnd+`shutdown contract"}`)
}

func TestRand(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	rnd := `rnd` + crypto.RandSeq(4)

	form := url.Values{`Value`: {`contract ` + rnd + ` {
		action {
			var result i int
			i = 3
			while i < 15 {
				var rnd int
				rnd = Random(0, 3*i)
				result = result + rnd
				i=i+1
			}
			$result = result
		}
	}`}, `Conditions`: {`true`}, `ApplicationId`: {`1`}}
	assert.NoError(t, postTx(`NewContract`, &form))
	_, val1, err := postTxResult(rnd, &url.Values{})
	assert.NoError(t, err)
	_, val2, err := postTxResult(rnd, &url.Values{})
	assert.NoError(t, err)
	// val1 == val2 for seed = blockId % 1
	if val1 != val2 {
		t.Errorf(`%s!=%s`, val1, val2)
	}
}
func TestKillNode(t *testing.T) {
	require.NoError(t, keyLogin(1))
	form := url.Values{"Name": {`MyTestContract1`}, "Value": {`contract MyTestContract1 {action {}}`},
		"ApplicationId": {`1`}, "Conditions": {`true`}, "nowait": {`true`}}
	require.NoError(t, postTx(`NewContract`, &form))
	require.NoError(t, postTx("Kill", &url.Values{"nowait": {`true`}}))
}

func TestLoopCondExt(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	rnd := `rnd` + crypto.RandSeq(4)

	form := url.Values{`Value`: {`contract ` + rnd + `1 {
		conditions {

		}
	}`}, `Conditions`: {`true`}, `ApplicationId`: {`1`}}
	err := postTx(`NewContract`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{`Value`: {`contract ` + rnd + `2 {
		conditions {
			ContractConditions("` + rnd + `1")
		}
	}`}, `Conditions`: {`true`}, `ApplicationId`: {`1`}}
	err = postTx(`NewContract`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	var ret getContractResult
	err = sendGet(`contract/`+rnd+`1`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	sid := ret.TableID
	form = url.Values{`Value`: {`contract ` + rnd + `1 {
		conditions {
			ContractConditions("` + rnd + `2")
		}
	}`}, `Id`: {sid}, `Conditions`: {`true`}, `ApplicationId`: {`1`}}
	err = postTx(`EditContract`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	err = postTx(rnd+`2`, &url.Values{})
	if err != nil {
		t.Error(err)
		return
	}
}
