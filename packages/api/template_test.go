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
	"time"

	"github.com/AplaProject/go-apla/packages/crypto"
)

type tplItem struct {
	input string
	want  string
}

type tplList []tplItem

func TestAPI(t *testing.T) {
	var ret contentResult

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	err := sendPost(`content/page/default_page`, &url.Values{}, &ret)
	if err != nil {
		t.Error(err)
		return
	}

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
	err = sendPost(`content/page/mypage`, &url.Values{}, &ret)
	if err != nil && err.Error() != `404 {"error": "E_NOTFOUND", "msg": "Page not found" }` {
		t.Error(err)
		return
	}
	err = sendPost(`content/menu/default_menu`, &url.Values{}, &ret)
	if err != nil {
		t.Error(err)
		return
	}
}

var forTest = tplList{
	{`SetVar(Name: vDateNow, Value: Now("YYYY-MM-DD HH:MI")) 
		SetVar(Name: vStartDate, Value: DateTime(DateTime: #vDateNow#, Format: "YYYY-MM-DD HH:MI"))
		SetVar(Name: vCmpStartDate, Value: CmpTime(#vStartDate#,#vDateNow#))
		Span(#vCmpStartDate#)`,
		`[{"tag":"span","children":[{"tag":"text","text":"0"}]}]`},
	{`Input(Type: text, Value: OK Now(YY)+Strong(Ooops))`,
		`[{"tag":"input","attr":{"type":"text","value":"OK 17+"}}]`},
	{`Button(Body: LangRes(save), Class: btn btn-primary, Contract: EditProfile, 
		Page:members_list,).Alert(Text: $want_save_changes$, 
		ConfirmButton: $yes$, CancelButton: $no$, Icon: question)`,
		`[{"tag":"button","attr":{"alert":{"cancelbutton":"$no$","confirmbutton":"$yes$","icon":"question","text":"$want_save_changes$"},"class":"btn btn-primary","contract":"EditProfile","page":"members_list"},"children":[{"tag":"text","text":"save"}]}]`},
	{`Simple Strong(bold text)`,
		`[{"tag":"text","text":"Simple "},{"tag":"strong","children":[{"tag":"text","text":"bold text"}]}]`},
	{`EcosysParam(gender, Source: mygender)`,
		`[{"tag":"data","attr":{"columns":["id","name"],"data":[["1",""]],"source":"mygender","types":["text","text"]}}]`},
	{`EcosysParam(new_table)`,
		`[{"tag":"text","text":"ContractConditions(\u0026#34;MainCondition\u0026#34;)"}]`},
	{`DBFind(pages,mypage).Columns("id,name,menu").Order(id).Vars(my)Strong(#my_menu#)`,
		`[{"tag":"dbfind","attr":{"columns":["id","name","menu"],"data":[["1","default_page","default_menu"]],"name":"pages","order":"id","source":"mypage","types":["text","text","text"]}},{"tag":"strong","children":[{"tag":"text","text":"default_menu"}]}]`},
}

func TestImage(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	name := randName(`tbl`)
	form := url.Values{"Name": {name}, "Columns": {`[{"name":"name","type":"varchar", "index": "1", 
	  "conditions":"true"},
	{"name":"image", "type":"text","index": "0", "conditions":"true"}]`},
		"Permissions": {`{"insert": "true", "update" : "true", "new_column": "true"}`}}
	err := postTx(`NewTable`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"Name": {name}, "Value": {`contract ` + name + ` {
		 data {
			 Image string
		 }
		 action {
			 DBInsert("` + name + `", "name,image", "myimage", $Image)
		 }
		}`},
		"Conditions": {`true`}}
	if err := postTx(`NewContract`, &form); err != nil {
		t.Error(err)
		return
	}
	var mydata string

	mydata = crypto.RandSeq(300000)
	err = postTx(name, &url.Values{`Image`: {mydata}})
	if err != nil {
		t.Error(err)
		return
	}
	var ret contentResult
	template := `DBFind(Name: ` + name + `, Source: srcimage).Custom(custom_image){
        Div(Body: Image(Src: "#image#").Style(width: 100px;  border: 1px solid #5A5D63 ;))
	}
	Table(Source: srcimage, Columns: "Name=name, Image=#custom_image#")
	`
	start := time.Now()
	err = sendPost(`content`, &url.Values{`template`: {template}}, &ret)
	duration := time.Since(start)
	if err != nil {
		t.Error(err)
		return
	}
	if int(duration.Seconds()) > 0 {
		t.Errorf(`Too much time for template parsing`)
		return
	}
}
