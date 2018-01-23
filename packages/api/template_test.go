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
	"crypto/md5"
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
	{`Data(Source: mysrc, Columns: "startdate,enddate", Data:
		2017-12-10 10:11,2017-12-12 12:13
		2017-12-17 16:17,2017-12-15 14:15
	).Custom(custom_id){
		SetVar(Name: vStartDate, Value: DateTime(DateTime: #startdate#, Format: "YYYY-MM-DD HH:MI"))
		SetVar(Name: vEndDate, Value: DateTime(DateTime: #enddate#, Format: "YYYY-MM-DD HH:MI"))
		SetVar(Name: vCmpDate, Value: CmpTime(#vStartDate#,#vEndDate#)) 
		P(Body: #vStartDate# #vEndDate# #vCmpDate#)
	}.Custom(custom_name){
		P(Body: #vStartDate# #vEndDate# #vCmpDate#)
	}`,
		`[{"tag":"data","attr":{"columns":["startdate","enddate","custom_id","custom_name"],"data":[["2017-12-10 10:11","2017-12-12 12:13","[{"tag":"p","children":[{"tag":"text","text":"2017-12-10 10:11 2017-12-12 12:13 -1"}]}]","[{"tag":"p","children":[{"tag":"text","text":"2017-12-10 10:11 2017-12-12 12:13 -1"}]}]"],["2017-12-17 16:17","2017-12-15 14:15","[{"tag":"p","children":[{"tag":"text","text":"2017-12-17 16:17 2017-12-15 14:15 1"}]}]","[{"tag":"p","children":[{"tag":"text","text":"2017-12-17 16:17 2017-12-15 14:15 1"}]}]"]],"source":"mysrc","types":["text","text","tags","tags"]}}]`},
	{`Strong(SysParam(commission_size))`,
		`[{"tag":"strong","children":[{"tag":"text","text":"3"}]}]`},
	{`SetVar(Name: vDateNow, Value: Now("YYYY-MM-DD HH:MI")) 
		SetVar(Name: simple, Value: TestFunc(my value)) 
		SetVar(Name: vStartDate, Value: DateTime(DateTime: #vDateNow#, Format: "YYYY-MM-DD HH:MI"))
		SetVar(Name: vCmpStartDate, Value: CmpTime(#vStartDate#,#vDateNow#))
		Span(#vCmpStartDate# #simple#)`,
		`[{"tag":"span","children":[{"tag":"text","text":"0 TestFunc(my value)"}]}]`},
	{`Input(Type: text, Value: OK Now(YY)+Strong(Ooops))`,
		`[{"tag":"input","attr":{"type":"text","value":"OK 18+"}}]`},
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

func TestLinkData(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	name := randName(`tbl`)
	form := url.Values{
		"Name": {name},
		"Columns": {`[
			{"name":"name","type":"varchar", "index": "1", "conditions":"true"},
			{"name":"image", "type":"bytea","index": "0", "conditions":"true"},
			{"name":"long_text", "type":"text", "index":"0", "conditions":"true"},
			{"name":"short_text", "type":"varchar", "index":"0", "conditions":"true"}
		]`},
		"Permissions": {`{"insert": "true", "update" : "true", "new_column": "true"}`},
	}
	err := postTx(`NewTable`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"Name": {name}, "Value": {`contract ` + name + ` {
		 data {
			 Image string
			 LongText string
			 ShortText string
		 }
		 action {
			 DBInsert("` + name + `", "name,image,long_text,short_text", "myimage", $Image, $LongText, $ShortText)
		 }
		}`},
		"Conditions": {`true`}}
	if err := postTx(`NewContract`, &form); err != nil {
		t.Error(err)
		return
	}

	mydata := `data:image/png;base64,` + crypto.RandSeq(30000)
	shortText := crypto.RandSeq(30)
	longText := crypto.RandSeq(100)

	err = postTx(name, &url.Values{
		"Image":     {mydata},
		"ShortText": {shortText},
		"LongText":  {longText},
	})
	if err != nil {
		t.Error(err)
		return
	}
	var ret contentResult
	template := `DBFind(Name: ` + name + `, Source: srcimage).Cutoff("short_text,long_text").Custom(custom_image){
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
	mydata = `data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADIAAAAyCAIAAACRXR/mAAAACXBIWXMAAAsTAAALEwEAmpwYAAAARklEQVRYw+3OMQ0AIBAEwQOzaCLBBQZfAd0XFLMCNjOyb1o7q2Ey82VYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYrwqjmwKzLUjCbwAAAABJRU5ErkJggg==`
	err = postTx(name, &url.Values{
		"Image":     {mydata},
		"ShortText": {shortText},
		"LongText":  {longText},
	})

	template = `Div(Class: list-group-item){
		Div(panel-body){
		   DBFind("` + name + `", mysrc).Columns("id,name,image,short_text,long_text").Cutoff("short_text,long_text").WhereId(2).Custom(leftImg){
			   Image(Src: "#image#")
		   }
		   }
		   Table(mysrc,"Image=leftImg")
		}
	 Form(){
	   ImageInput(Name: img, Width: 400, Ratio: 2/1)
				Button(Body: Add, Contract: UploadImage){ Upload! }
	  }`
	err = sendPost(`content`, &url.Values{`template`: {template}}, &ret)
	if err != nil {
		t.Error(err)
		return
	}

	hashImage := fmt.Sprintf("%x", md5.Sum([]byte(mydata)))
	linkImage := fmt.Sprintf("/data/1_%s/2/image/%s", name, hashImage)
	linkLongText := fmt.Sprintf("/data/1_%s/2/long_text/%x", name, md5.Sum([]byte(longText)))

	want := `[{"tag":"div","attr":{"class":"list-group-item"},"children":[{"tag":"div","attr":{"class":"panel-body"},"children":[{"tag":"dbfind","attr":{"columns":["id","name","image","short_text","long_text","leftImg"],"cutoff":"short_text,long_text","data":[["2","myimage","{"link":"` + linkImage + `","title":"` + hashImage + `"}","{"link":"","title":"` + shortText + `"}","{"link":"` + linkLongText + `","title":"` + longText[0:32] + `"}","[{"tag":"image","attr":{"src":"` + linkImage + `"}}]"]],"name":"` + name + `","source":"mysrc","types":["text","text","blob","long_text","long_text","tags"],"whereid":"2"}}]},{"tag":"table","attr":{"columns":[{"Name":"leftImg","Title":"Image"}],"source":"mysrc"}}]},{"tag":"form","children":[{"tag":"imageinput","attr":{"name":"img","ratio":"2/1","width":"400"}},{"tag":"button","attr":{"contract":"UploadImage"},"children":[{"tag":"text","text":"Upload!"}]}]}]`
	if RawToString(ret.Tree) != want {
		t.Errorf("Wrong image tree %s != %s", RawToString(ret.Tree), want)
	}

	data, err := sendRawRequest("GET", linkLongText, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if string(data) != longText {
		t.Errorf("Wrong text %s", data)
	}
}
