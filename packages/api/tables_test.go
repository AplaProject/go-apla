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

func TestTables(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	var ret tablesResult
	err := sendGet(`tables`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if int64(ret.Count) < 7 {
		t.Error(fmt.Errorf(`The number of tables %d < 7`, ret.Count))
		return
	}
}

func TestTable(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	var ret tableResult
	err := sendGet(`table/keys`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if len(ret.Columns) == 0 {
		t.Error(err)
		return
	}
	err = sendGet(`table/contracts`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestJSONTable(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	name := randName(`json`)
	form := url.Values{"Name": {name}, "Columns": {`[{"name":"MyName","type":"varchar", "index": "0", 
	  "conditions":"true"}, {"name":"Doc", "type":"json","index": "0", "conditions":"true"}]`},
		"Permissions": {`{"insert": "true", "update" : "true", "new_column": "true"}`}}
	err := postTx(`NewTable`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	checkGet := func(want string) {
		_, msg, err := postTxResult(name+`Get`, &url.Values{"Id": {`2`}})
		if err != nil {
			t.Error(err)
			return
		}
		if msg != want {
			t.Error(`wrong answer`, msg)
		}
	}

	form = url.Values{"Name": {name}, "Value": {`contract ` + name + ` {
		action { 
			var ret1, ret2 int
			ret1 = DBInsert("` + name + `", "MyName,Doc", "test", "{\"type\": \"0\"}")
			var mydoc map
			mydoc["type"] = "document"
			mydoc["ind"] = 2
			mydoc["check"] = "99"
			mydoc["doc"] = "Some text."
			ret2 = DBInsert("` + name + `", "MyName,Doc", "test2", mydoc)
			DBInsert("` + name + `", "MyName,Doc", "test3", "{\"title\": {\"name\":\"Test att\",\"text\":\"low\"}}")
		}}`},
		"Conditions": {`ContractConditions("MainCondition")`}}
	err = postTx("NewContract", &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"Name": {name}, "Value": {`contract ` + name + `Get {
			data {
				Id int
			}
			action {
				var ret map
				var list array
				var out tmp string
				ret = DBFind("` + name + `").Columns("Myname,doc,Doc->Ind").WhereId($Id).Row()
				out = ret["doc.ind"]
				out = out + DBFind("` + name + `").Columns("myname,doc->Type").WhereId($Id).One("Doc->type")
				list = DBFind("` + name + `").Columns("Myname,doc,Doc->Ind").Where("Doc->ind = ?", "101")
				out = out + Str(Len(list))
				tmp = DBFind("` + name + `").Columns("doc->title->name").WhereId(3).One("doc->title->name")
				$result = out + Str(DBFind("` + name + `").WhereId($Id).One("doc->check")) + tmp
			}
		}`},
		"Conditions": {`ContractConditions("MainCondition")`}}
	err = postTx("NewContract", &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"Name": {name}, "Value": {`contract ` + name + `Upd {
		action {
			DBUpdate("` + name + `", 1, "Doc", "{\"type\": \"doc\", \"ind\": \"3\", \"check\": \"33\"}")
			var mydoc map
			mydoc["type"] = "doc"
			mydoc["doc"] = "Some test text."
			DBUpdate("` + name + `", 2, "myname,Doc", "test3", mydoc)
		}}`},
		"Conditions": {`ContractConditions("MainCondition")`}}
	err = postTx("NewContract", &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"Name": {name}, "Value": {`contract ` + name + `UpdOne {
			data {
				Type int
			}
			action {
				DBUpdate("` + name + `", 1, "myname,Doc->Ind,Doc->type", "New name", 
					      $Type, "new\"doc\" val")
				DBUpdate("` + name + `", 2, "myname,Doc->Ind,Doc->type", "New name", 
						$Type, "new\"doc\"")
				DBUpdate("` + name + `", 3, "doc->flag,doc->sub", "Flag", 100)
				DBUpdate("` + name + `", 3, "doc->temp", "Temp")
		  }}
		`},
		"Conditions": {`ContractConditions("MainCondition")`}}
	err = postTx("NewContract", &form)
	if err != nil {
		t.Error(err)
		return
	}
	err = postTx(name, &url.Values{})
	if err != nil {
		t.Error(err)
		return
	}
	checkGet(`2document099Test att`)

	err = postTx(name+`Upd`, &url.Values{})
	if err != nil {
		t.Error(err)
		return
	}
	checkGet(`doc0Test att`)

	err = postTx(name+`UpdOne`, &url.Values{"Type": {"101"}})
	if err != nil {
		t.Error(err)
		return
	}
	checkGet(`101new"doc"2Test att`)

	form = url.Values{"Name": {`res` + name}, "Value": {`contract res` + name + ` {
		data {
			Id int
		}
		action { 
			$result = DBFind("contracts").WhereId($Id).Row()
		}}`},
		"Conditions": {`ContractConditions("MainCondition")`}}
	err = postTx("NewContract", &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"Name": {`run` + name}, "Value": {`contract run` + name + ` {
		action { 
			$temp = res` + name + `("Id",10)
			$result = $temp["id"]
		}}`},
		"Conditions": {`ContractConditions("MainCondition")`}}
	err = postTx("NewContract", &form)
	if err != nil {
		t.Error(err)
		return
	}
	_, msg, err := postTxResult(`run`+name, &url.Values{})
	if err != nil {
		t.Error(err)
		return
	}
	if msg != `10` {
		t.Error(`wrong answer`, msg)
	}

	forTest := tplList{
		{`DBFind(` + name + `,my).Columns("id,doc->title->name").WhereId(3)`,
			`[{"tag":"dbfind","attr":{"columns":["id","doc.title.name"],"data":[["3","Test att"]],"name":"` + name + `","source":"my","types":["text","text"],"whereid":"3"}}]`},
		{`DBFind(` + name + `,my).Columns("doc").WhereId(3)`,
			`[{"tag":"dbfind","attr":{"columns":["doc","id"],"data":[["{"sub": "100", "flag": "Flag", "temp": "Temp"}","3"]],"name":"` + name + `","source":"my","types":["text","text"],"whereid":"3"}}]`},
		{`DBFind(` + name + `,my).Columns("id,doc,doc->type").Where(doc->ind='101' and doc->check='33')`,
			`[{"tag":"dbfind","attr":{"columns":["id","doc","doc.type"],"data":[["1","{"ind": "101", "type": "new\\"doc\\" val", "check": "33"}","new"doc" val"]],"name":"` + name + `","source":"my","types":["text","text","text"],"where":"doc-\u003eind='101' and doc-\u003echeck='33'"}}]`},
		{`DBFind(` + name + `,my).Columns("id,doc,doc->type").WhereId(2).Vars(my)
			Span(#my_id##my_doc_type#)`,
			`[{"tag":"dbfind","attr":{"columns":["id","doc","doc.type"],"data":[["2","{"doc": "Some test text.", "ind": "101", "type": "new\\"doc\\""}","new"doc""]],"name":"` + name + `","source":"my","types":["text","text","text"],"whereid":"2"}},{"tag":"span","children":[{"tag":"text","text":"2new"doc""}]}]`},
		{`DBFind(` + name + `,my).Columns("id,doc->type").WhereId(2)`,
			`[{"tag":"dbfind","attr":{"columns":["id","doc.type"],"data":[["2","new"doc""]],"name":"` + name + `","source":"my","types":["text","text"],"whereid":"2"}}]`},
		{`DBFind(` + name + `,my).Columns("doc->type").Custom(mytype, OK:#doc.type#)`,
			`[{"tag":"dbfind","attr":{"columns":["doc.type","id","mytype"],"data":[["new"doc" val","1","[{"tag":"text","text":"OK:new\\u0026#34;doc\\u0026#34; val"}]"],["new"doc"","2","[{"tag":"text","text":"OK:new\\u0026#34;doc\\u0026#34;"}]"]],"name":"` +
				name + `","source":"my","types":["text","text","tags"]}}]`},
	}
	var ret contentResult
	for _, item := range forTest {
		err := sendPost(`content`, &url.Values{`template`: {item.input}}, &ret)
		if err != nil {
			t.Error(err)
			return
		}
		if RawToString(ret.Tree) != item.want {
			t.Error(fmt.Errorf(`wrong tree %s != %s`, RawToString(ret.Tree), item.want))
			return
		}
	}
}
