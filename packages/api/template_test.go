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
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/stretchr/testify/assert"
)

type tplItem struct {
	input string
	want  string
}

type tplList []tplItem

func TestAPI(t *testing.T) {
	var ret contentResult

	requireLogin(t, 1)
	require.NoError(t, sendPost(`content/page/default_page`, &url.Values{}, &ret))

	var retHash hashResult
	require.NoError(t, sendPost(`content/hash/default_page`, &url.Values{}, &retHash))
	require.Len(t, retHash.Hash, 64, `wrong hash `, retHash.Hash)

	for _, item := range forTest {
		require.NoError(t, sendPost(`content`, &url.Values{`template`: {item.input}}, &ret))
		require.Equalf(t, item.want, RawToString(ret.Tree), `wrong tree %s != %s`, RawToString(ret.Tree), item.want)
	}

	require.Error(t, sendPost(`content/page/mypage`, &url.Values{}, &ret), `404 {"error": "E_NOTFOUND", "msg": "Page not found" }`)
	require.NoError(t, sendPost(`content/menu/default_menu`, &url.Values{}, &ret))
}

var forTest = tplList{
	{`SetVar(coltype, GetColumnType(members, member_name))Div(){#coltype#GetColumnType(none,none)GetColumnType()}`, `[{"tag":"div","children":[{"tag":"text","text":"varchar"}]}]`},
	{`DBFind(parameters, src_par).Columns("id").Order(id).Where("id >= 1 and id <= 3").Count(count)Span(#count#)`,
		`[{"tag":"dbfind","attr":{"columns":["id"],"count":"3","data":[["1"],["2"],["3"]],"name":"parameters","order":"id","source":"src_par","types":["text"],"where":"id \u003e= 1 and id \u003c= 3"}},{"tag":"span","children":[{"tag":"text","text":"3"}]}]`},
	{`SetVar(coltype, GetColumnType(members, member_name))Div(){#coltype#GetColumnType(none,none)GetColumnType()}`, `[{"tag":"div","children":[{"tag":"text","text":"varchar"}]}]`},
	{`SetVar(where)DBFind(contracts, src).Columns(id).Order(id).Limit(3).Custom(a){SetVar(where, #where# #id#)}
	Div(){Table(src, "=x")}Div(){Table(src)}Div(){#where#}`,
		`[{"tag":"dbfind","attr":{"columns":["id","a"],"data":[["1","null"],["2","null"],["3","null"]],"limit":"3","name":"contracts","order":"id","source":"src","types":["text","tags"]}},{"tag":"div","children":[{"tag":"table","attr":{"columns":[{"Name":"x","Title":""}],"source":"src"}}]},{"tag":"div","children":[{"tag":"table","attr":{"source":"src"}}]},{"tag":"div","children":[{"tag":"text","text":" 1 2 3"}]}]`},
	{`If(#isMobile#){Span(Mobile)}.Else{Span(Desktop)}`,
		`[{"tag":"span","children":[{"tag":"text","text":"Desktop"}]}]`},
	{`DBFind(contracts, src_contracts).Columns("id").Order(id).Limit(2).Offset(10)`,
		`[{"tag":"dbfind","attr":{"columns":["id"],"data":[["11"],["12"]],"limit":"2","name":"contracts","offset":"10","order":"id","source":"src_contracts","types":["text"]}}]`},
	{`DBFind(contracts, src_pos).Columns(id).Where("id >= 1 and id <= 3")
		ForList(src_pos, Index: index){
			Div(list-group-item) {
				DBFind(parameters, src_hol).Columns(id).Where("id=#id#").Vars("ret")
				SetVar(qq, #ret_id#)
				Div(Body: #index# ForList=#id# DBFind=#ret_id# SetVar=#qq#)  
			}
		}`, `[{"tag":"dbfind","attr":{"columns":["id"],"data":[["1"],["2"],["3"]],"name":"contracts","source":"src_pos","types":["text"],"where":"id \u003e= 1 and id \u003c= 3"}},{"tag":"forlist","attr":{"index":"index","source":"src_pos"},"children":[{"tag":"div","attr":{"class":"list-group-item"},"children":[{"tag":"dbfind","attr":{"columns":["id"],"data":[["1"]],"name":"parameters","source":"src_hol","types":["text"],"where":"id=1"}},{"tag":"div","children":[{"tag":"text","text":"1 ForList=1 DBFind=1 SetVar=1"}]}]},{"tag":"div","attr":{"class":"list-group-item"},"children":[{"tag":"dbfind","attr":{"columns":["id"],"data":[["2"]],"name":"parameters","source":"src_hol","types":["text"],"where":"id=2"}},{"tag":"div","children":[{"tag":"text","text":"2 ForList=2 DBFind=2 SetVar=2"}]}]},{"tag":"div","attr":{"class":"list-group-item"},"children":[{"tag":"dbfind","attr":{"columns":["id"],"data":[["3"]],"name":"parameters","source":"src_hol","types":["text"],"where":"id=3"}},{"tag":"div","children":[{"tag":"text","text":"3 ForList=3 DBFind=3 SetVar=3"}]}]}]}]`},
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
	{`Button(Body: LangRes(savex), Class: btn btn-primary, Contract: EditProfile, 
		Page:members_list,).Alert(Text: $want_save_changesx$, 
		ConfirmButton: $yesx$, CancelButton: $nox$, Icon: question)`,
		`[{"tag":"button","attr":{"alert":{"cancelbutton":"$nox$","confirmbutton":"$yesx$","icon":"question","text":"$want_save_changesx$"},"class":"btn btn-primary","contract":"EditProfile","page":"members_list"},"children":[{"tag":"text","text":"savex"}]}]`},
	{`Simple Strong(bold text)`,
		`[{"tag":"text","text":"Simple "},{"tag":"strong","children":[{"tag":"text","text":"bold text"}]}]`},
	{`EcosysParam(gender, Source: mygender)`,
		`[{"tag":"data","attr":{"columns":["id","name"],"data":[["1",""]],"source":"mygender","types":["text","text"]}}]`},
	{`EcosysParam(new_table)`,
		`[{"tag":"text","text":"ContractConditions("MainCondition")"}]`},
	{`DBFind(pages,mypage).Columns("id,name,menu").Order(id).Vars(my)Strong(#my_menu#)`,
		`[{"tag":"dbfind","attr":{"columns":["id","name","menu"],"data":[["1","default_page","default_menu"]],"name":"pages","order":"id","source":"mypage","types":["text","text","text"]}},{"tag":"strong","children":[{"tag":"text","text":"default_menu"}]}]`},
	{`SetVar(varZero, 0) If(#varZero#>0) { the varZero should be hidden }
		SetVar(varNotZero, 1) If(#varNotZero#>0) { the varNotZero should be visible }
		If(#varUndefined#>0) { the varUndefined should be hidden }`,
		`[{"tag":"text","text":"the varNotZero should be visible"}]`},
	{`Address(EcosysParam(founder_account))+EcosysParam(founder_account)`,
		`[{"tag":"text","text":"1651-3553-1389-2023-2108"},{"tag":"text","text":"+-1933190934789319508"}]`},
}

func TestMobile(t *testing.T) {
	var ret contentResult
	gMobile = true
	requireLogin(t, 1)
	require.NoError(t, sendPost(`content`, &url.Values{`template`: {`If(#isMobile#){Span(Mobile)}.Else{Span(Desktop)}`}}, &ret))
	require.Equalf(t, `[{"tag":"span","children":[{"tag":"text","text":"Mobile"}]}]`, RawToString(ret.Tree), `wrong mobile tree %s`, RawToString(ret.Tree))
}

func TestCutoff(t *testing.T) {
	requireLogin(t, 1)
	name := randName(`tbl`)
	form := url.Values{
		"Name": {name},
		"Columns": {`[
			{"name":"name","type":"varchar", "index": "1", "conditions":"true"},
			{"name":"long_text", "type":"text", "index":"0", "conditions":"true"},
			{"name":"short_text", "type":"varchar", "index":"0", "conditions":"true"}
			]`},
		"Permissions": {`{"insert": "true", "update" : "true", "new_column": "true"}`},
	}
	require.NoError(t, postTx(`NewTable`, &form))

	form = url.Values{
		"Name": {name},
		"Value": {`
			contract ` + name + ` {
				data {
					LongText string
					ShortText string
				}
				action {
					DBInsert("` + name + `", "name,long_text,short_text", "test", $LongText, $ShortText)
				}
			}
		`},
		"Conditions": {`true`},
	}
	require.NoError(t, postTx(`NewContract`, &form))

	shortText := crypto.RandSeq(30)
	longText := crypto.RandSeq(100)

	require.NoError(t, postTx(name, &url.Values{
		"ShortText": {shortText},
		"LongText":  {longText},
	}))

	var ret contentResult
	template := `DBFind(Name: ` + name + `, Source: mysrc).Cutoff("short_text,long_text")`
	start := time.Now()
	duration := time.Since(start)
	require.NoError(t, sendPost(`content`, &url.Values{`template`: {template}}, &ret))
	require.Equal(t, 0, int(duration.Seconds()), `Too much time for template parsing`)

	require.NoError(t, postTx(name, &url.Values{
		"ShortText": {shortText},
		"LongText":  {longText},
	}))

	template = `DBFind("` + name + `", mysrc).Columns("id,name,short_text,long_text").Cutoff("short_text,long_text").WhereId(2).Vars(prefix)`
	require.NoError(t, sendPost(`content`, &url.Values{`template`: {template}}, &ret))

	linkLongText := fmt.Sprintf("/data/1_%s/2/long_text/%x", name, md5.Sum([]byte(longText)))

	want := `[{"tag":"dbfind","attr":{"columns":["id","name","short_text","long_text"],"cutoff":"short_text,long_text","data":[["2","test","{"link":"","title":"` + shortText + `"}","{"link":"` + linkLongText + `","title":"` + longText[:32] + `"}"]],"name":"` + name + `","source":"mysrc","types":["text","text","long_text","long_text"],"whereid":"2"}}]`
	require.Equalf(t, want, RawToString(ret.Tree), "Wrong image tree %s != %s", RawToString(ret.Tree), want)

	data, err := sendRawRequest("GET", linkLongText, nil)
	require.NoError(t, err)
	require.Equalf(t, longText, string(data), "Wrong text %s", data)
}

var imageData = `iVBORw0KGgoAAAANSUhEUgAAADIAAAAyCAIAAACRXR/mAAAACXBIWXMAAAsTAAALEwEAmpwYAAAARklEQVRYw+3OMQ0AIBAEwQOzaCLBBQZfAd0XFLMCNjOyb1o7q2Ey82VYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYrwqjmwKzLUjCbwAAAABJRU5ErkJggg==`

func TestBinary(t *testing.T) {
	requireLogin(t, 1)

	params := map[string]string{
		"AppID":    "1",
		"MemberID": "1",
		"Name":     "file",
	}

	data, err := base64.StdEncoding.DecodeString(imageData)
	assert.NoError(t, err)

	files := map[string][]byte{
		"Data": data,
	}

	_, _, err = postTxMultipart("UploadBinary", params, files)
	assert.NoError(t, err)

	hashImage := fmt.Sprintf("%x", md5.Sum(data))

	cases := []struct {
		source string
		result string
	}{
		{
			`Image(Src: Binary(Name: file, AppID: 1, MemberID: 1))`,
			`\[{"tag":"image","attr":{"src":"/data/1_binaries/\d+/data/` + hashImage + `"}}\]`,
		},
		{
			`DBFind(Name: binaries, Src: mysrc).Where("app_id=1 AND member_id = 1 AND name = 'file'").Custom(img){Image(Src: #data#)}Table(mysrc, "Image=img")`,
			`\[{"tag":"dbfind","attr":{"columns":\["id","app_id","member_id","name","data","hash","mime_type","img"\],"data":\[\["\d+","1","1","file","{\\"link\\":\\"/data/1_binaries/\d+/data/` + hashImage + `\\",\\"title\\":\\"` + hashImage + `\\"}","` + hashImage + `","application/octet-stream","\[{\\"tag\\":\\"image\\",\\"attr\\":{\\"src\\":\\"/data/1_binaries/\d+/data/` + hashImage + `\\"}}\]"\]\],"name":"binaries","source":"Src: mysrc","types":\["text","text","text","text","blob","text","text","tags"\],"where":"app_id=1 AND member_id = 1 AND name = 'file'"}},{"tag":"table","attr":{"columns":\[{"Name":"img","Title":"Image"}\],"source":"mysrc"}}\]`,
		},
		{
			`DBFind(Name: binaries, Src: mysrc).Where("app_id=1 AND member_id = 1 AND name = 'file'").Vars(prefix)Image(Src: "#prefix_data#")`,
			`\[{"tag":"dbfind","attr":{"columns":\["id","app_id","member_id","name","data","hash","mime_type"\],"data":\[\["\d+","1","1","file","{\\"link\\":\\"/data/1_binaries/\d+/data/` + hashImage + `\\",\\"title\\":\\"` + hashImage + `\\"}","` + hashImage + `","application/octet-stream"\]\],"name":"binaries","source":"Src: mysrc","types":\["text","text","text","text","blob","text","text"\],"where":"app_id=1 AND member_id = 1 AND name = 'file'"}},{"tag":"image","attr":{"src":"{\\"link\\":\\"/data/1_binaries/\d+/data/` + hashImage + `\\",\\"title\\":\\"` + hashImage + `\\"}"}}\]`,
		},
	}

	for _, v := range cases {
		var ret contentResult
		require.NoError(t, sendPost(`content`, &url.Values{`template`: {v.source}}, &ret))
		assert.Regexp(t, v.result, string(ret.Tree))
	}
}

func TestStringToBinary(t *testing.T) {
	requireLogin(t, 1)

	contract := randName("binary")
	content := randName("content")

	form := url.Values{
		"Value": {`
			contract ` + contract + ` {
				data {
					Content string
				}
				conditions {}
				action {
					UploadBinary("Name,AppID,Data,DataMimeType", "test", 1, StringToBytes($Content), "text/plain")
				}
			}
		`},
		"Conditions": {"true"},
	}
	assert.NoError(t, postTx("NewContract", &form))

	form = url.Values{"Content": {content}}
	assert.NoError(t, postTx(contract, &form))

	form = url.Values{
		"template": {`SetVar(link, Binary(Name: test, AppID: 1)) #link#`},
	}
	var ret struct {
		Tree []struct {
			Link string `json:"text"`
		} `json:"tree"`
	}
	require.NoError(t, sendPost(`content`, &form, &ret))

	data, err := sendRawRequest("GET", strings.TrimSpace(ret.Tree[0].Link), nil)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
}
