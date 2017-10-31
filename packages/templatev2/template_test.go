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

package templatev2

import (
	"testing"
)

type tplItem struct {
	input string
	want  string
}

type tplList []tplItem

func TestJSON(t *testing.T) {
	vars := make(map[string]string)
	for _, item := range forTest {
		templ := Template2JSON(item.input, false, &vars)
		if string(templ) != item.want {
			t.Errorf(`wrong json %s != %s`, templ, item.want)
			return
		}
	}
}

var forTest = tplList{
	{`LinkPage(My page,mypage)`,
		`[{"tag":"linkpage","attr":{"page":"mypage"},"children":[{"tag":"text","text":"My page"}]}]`},
	{`Image(/images/myimage.jpg,My photo,myclass)`,
		`[{"tag":"image","attr":{"alt":"My photo","class":"myclass","src":"/images/myimage.jpg"}}]`},
	{`Data(mysrc,"id,name",
		"1",John Silver,2
		2,"Mark, Smith"
	)`,
		`[{"tag":"data","attr":{"columns":["id","name"],"data":[],"error":"line 2, column 0: wrong number of fields in line","source":"mysrc"}}]`},
	{`Select(mysrc,name,myclass)`,
		`[{"tag":"select","attr":{"class":"myclass","column":"name","source":"mysrc"}}]`},
	{`Data(mysrc,"id,name"){
		"1",John Silver
		2,"Mark, Smith"
		3,"Unknown ""Person"""
		}`,
		`[{"tag":"data","attr":{"columns":["id","name"],"data":[["1","John Silver"],["2","Mark, Smith"],["3","Unknown \"Person\""]],"source":"mysrc"}}]`},
	{`If(true) {OK}.Else {false} Div(){test} If(false, FALSE).ElseIf(0) { Skip }.ElseIf(1) {Else OK
		}.Else {Fourth}If(0).Else{ALL right}`,
		`[{"tag":"text","text":"OK"},{"tag":"div","children":[{"tag":"text","text":"test"}]},{"tag":"text","text":"Else OK"},{"tag":"text","text":"ALL right"}]`},
	{`Button(Contract: MyContract, Body:My Contract, Class: myclass, Params:"Name=myid,Id=i10,Value")`,
		`[{"tag":"button","attr":{"class":"myclass","contract":"MyContract","params":{"Id":"i10","Name":"myid","Value":"Value"}},"children":[{"tag":"text","text":"My Contract"}]}]`},
	{`Simple text +=<b>bold</b>`, `[{"tag":"text","text":"Simple text +=\u0026lt;b\u0026gt;bold\u0026lt;/b\u0026gt;"}]`},
	{`Div(myclass control, Content of the Div)`, `[{"tag":"div","attr":{"class":"myclass control"},"children":[{"tag":"text","text":"Content of the Div"}]}]`},
	{`input Div(myclass, Content Div(mypar) the Div)`,
		`[{"tag":"text","text":"input "},{"tag":"div","attr":{"class":"myclass"},"children":[{"tag":"text","text":"Content "},{"tag":"div","attr":{"class":"mypar"}},{"tag":"text","text":" the Div"}]}]`},
	{`Div(, Input(myid, form-control, Your name)Input(,,,text))`,
		`[{"tag":"div","children":[{"tag":"input","attr":{"class":"form-control","name":"myid","placeholder":"Your name"}},{"tag":"input","attr":{"type":"text"}}]}]`},
	{`Div(Class: mydiv1, Body:
			Div(Class: mydiv2,
				Div(Body:
					Input(Value: my default text))))`,
		`[{"tag":"div","attr":{"class":"mydiv1"},"children":[{"tag":"div","attr":{"class":"mydiv2"},"children":[{"tag":"div","children":[{"tag":"input","attr":{"value":"my default text"}}]}]}]}]`},
	{`P(Some Span(fake(text) Strong(very Em(important Label(news)))))`,
		`[{"tag":"p","children":[{"tag":"text","text":"Some "},{"tag":"span","children":[{"tag":"text","text":"fake(text) "},{"tag":"strong","children":[{"tag":"text","text":"very "},{"tag":"em","children":[{"tag":"text","text":"important "},{"tag":"label","children":[{"tag":"text","text":"news"}]}]}]}]}]}]`},
	{`Form(myclass, Input(myid)Button(Submit,default_page,myclass))`,
		`[{"tag":"form","attr":{"class":"myclass"},"children":[{"tag":"input","attr":{"name":"myid"}},{"tag":"button","attr":{"class":"myclass","page":"default_page"},"children":[{"tag":"text","text":"Submit"}]}]}]`},
	{`Button(My Contract,, myclass, NewEcosystem, "Name=myid,Id=i10,Value").Style( .btn {
		border: 10px 10px;
	})`,
		`[{"tag":"button","attr":{"class":"myclass","contract":"NewEcosystem","params":{"Id":"i10","Name":"myid","Value":"Value"},"style":".btn {\n\t\tborder: 10px 10px;\n\t}"},"children":[{"tag":"text","text":"My Contract"}]}]`},
	{`Div(myclass)Div().Style{
		.class {
			text-style: italic;
		}
	}
				Div()`,
		`[{"tag":"div","attr":{"class":"myclass"}},{"tag":"div","attr":{"style":".class {\n\t\t\ttext-style: italic;\n\t\t}"}},{"tag":"div"}]`},
	{`Div(myclass){Div()
		P(){
			Div(id){
				Label(My #text#,myl,forname)
			}
		}
	}`,
		`[{"tag":"div","attr":{"class":"myclass"},"children":[{"tag":"div"},{"tag":"p","children":[{"tag":"div","attr":{"class":"id"},"children":[{"tag":"label","attr":{"class":"myl","for":"forname"},"children":[{"tag":"text","text":"My #text#"}]}]}]}]}]`},
	{`SetVar(istrue, 1)If(GetVar(istrue),OK)If(GetVar(isfalse)){Skip}.Else{Span(Else OK)}`,
		`[{"tag":"text","text":"OK"},{"tag":"span","children":[{"tag":"text","text":"Else OK"}]}]`},
	{`If(false,First).ElseIf(0){Skip}.ElseIf(1){
		Second Span(If(text){item})
	}.ElseIf(true){Third}.Else{Fourth}`,
		`[{"tag":"text","text":"Second "},{"tag":"span","children":[{"tag":"text","text":"item"}]}]`},
	{`Button(Page: link){My Button}.Alert(ConfirmButton: ConfBtn, CancelButton: CancelBtn, 
		   Text: Alert text, Icon:myicon)`,
		`[{"tag":"button","attr":{"alert":{"cancelbutton":"CancelBtn","confirmbutton":"ConfBtn","icon":"myicon","text":"Alert text"},"page":"link"},"children":[{"tag":"text","text":"My Button"}]}]`},
	{`Input(myid, form-control, Your name).Validate(minLength: 6, maxLength: 20)
	InputErr(Name: myid, minLength: minLength error)`,
		`[{"tag":"input","attr":{"class":"form-control","name":"myid","placeholder":"Your name","validate":{"maxlength":"20","minlength":"6"}}},{"tag":"inputerr","attr":{"minlength":"minLength error","name":"myid"}}]`},
	{`MenuItem(Menu 1,page1)MenuGroup(SubMenu){
		MenuItem(Menu 2, page2)
		MenuItem(Page: page3, Title: Menu 3, Icon: person)
		}`,
		`[{"tag":"menuitem","attr":{"page":"page1","title":"Menu 1"}},{"tag":"menugroup","attr":{"title":"SubMenu"},"children":[{"tag":"menuitem","attr":{"page":"page2","title":"Menu 2"}},{"tag":"menuitem","attr":{"icon":"person","page":"page3","title":"Menu 3"}}]}]`},
	{`SetVar(testvalue, The, #n#, Value).(n, New).(param,"23")Span(Test value equals #testvalue#).(#param#)`,
		`[{"tag":"span","children":[{"tag":"text","text":"Test value equals The, New, Value"}]},{"tag":"span","children":[{"tag":"text","text":"23"}]}]`},
	{`SetVar(test, mytest).(empty,0)And(0,test,0)Or(0,#test#)Or(0, And(0,0))And(0,Or(0,my,while))
		And(1,#mytest#)Or(#empty#, And(#empty#, line))`,
		`[{"tag":"text","text":"010010"}]`},
	{`Address()Span(Address(-5728238900021))Address(3467347643873).(-6258391547979339691)`,
		`[{"tag":"text","text":"unknown address"},{"tag":"span","children":[{"tag":"text","text":"1844-6738-3454-7065-1595"}]},{"tag":"text","text":"0000-0003-4673-4764-38731218-8352-5257-3021-1925"}]`},
	{`Table(src, "id=ID,name,wallet=Wallet")`,
		`[{"tag":"table","attr":{"columns":{"id":"ID","name":"name","wallet":"Wallet"},"source":"src"}}]`},
	/*	{`Div(myclass, Include(test)Span(OK))`,
			`[{"tag":"include","attr":{"name":"myblock"}}]`},
			{`DBFind(1_keys).Columns(id,amount).WhereId(10).Limit(25).Custom(myid){
		 			Strong(#id#, text-muted)
		 			Div(,div element: #id#)
		 		 }.Custom(mybtn){Button(#amount#,mypage=#id#)}`,
			``},*/}

func TestFullJSON(t *testing.T) {
	vars := make(map[string]string)
	for _, item := range forFullTest {
		templ := Template2JSON(item.input, true, &vars)
		if string(templ) != item.want {
			t.Errorf(`wrong json %s != %s`, templ, item.want)
			return
		}
	}
}

var forFullTest = tplList{
	{`DBFind(parameters, mysrc).Columns("name,amount").Limit(10)Table(mysrc,"Name=name,Amount=amount").Style(.tbl {boder: 0px;})`,
		`[{"tag":"dbfind","attr":{"name":"parameters","source":"mysrc"},"tail":[{"tag":"columns","attr":{"columns":"name,amount"}},{"tag":"limit","attr":{"limit":"10"}}]},{"tag":"table","attr":{"columns":"Name=name,Amount=amount","source":"mysrc"},"tail":[{"tag":"style","attr":{"style":".tbl {boder: 0px;}"}}]}]`},
	{`Simple text +=<b>bold</b>`, `[{"tag":"text","text":"Simple text +=\u0026lt;b\u0026gt;bold\u0026lt;/b\u0026gt;"}]`},
	{`Div(myclass control, Content of the Div)`, `[{"tag":"div","attr":{"class":"myclass control"},"children":[{"tag":"text","text":"Content of the Div"}]}]`},
	{`If(true,OK)If(false){Skip}.Else{Span(Else OK)}`,
		`[{"tag":"if","attr":{"condition":"true"},"children":[{"tag":"text","text":"OK"}]},{"tag":"if","attr":{"condition":"false"},"children":[{"tag":"text","text":"Skip"}],"tail":[{"tag":"else","children":[{"tag":"span","children":[{"tag":"text","text":"Else OK"}]}]}]}]`},
	{`If(false,First).ElseIf(GetVar(my)){Skip}.ElseIf(1){
		Second
	}.ElseIf(true){Third}.Else{Fourth}`,
		`[{"tag":"if","attr":{"condition":"false"},"children":[{"tag":"text","text":"First"}],"tail":[{"tag":"elseif","attr":{"condition":"GetVar(my)"},"children":[{"tag":"text","text":"Skip"}]},{"tag":"elseif","attr":{"condition":"1"},"children":[{"tag":"text","text":"Second"}]},{"tag":"elseif","attr":{"condition":"true"},"children":[{"tag":"text","text":"Third"}]},{"tag":"else","children":[{"tag":"text","text":"Fourth"}]}]}]`},
	{`Button(Page: link){My Button}.Alert(ConfirmButton: ConfBtn, CancelButton: CancelBtn, 
			Text: Alert text, Icon:myicon)`,
		`[{"tag":"button","attr":{"page":"link"},"children":[{"tag":"text","text":"My Button"}],"tail":[{"tag":"alert","attr":{"cancelbutton":"CancelBtn","confirmbutton":"ConfBtn","icon":"myicon","text":"Alert text"}}]}]`},
	{`SetVar(testvalue, The new value).(n, param).Span(#testvalue#)`,
		`[{"tag":"setvar","attr":{"name":"testvalue","value":"The new value"}},{"tag":"setvar","attr":{"name":"n","value":"param"}},{"tag":"text","text":"."},{"tag":"span","children":[{"tag":"text","text":"#testvalue#"}]}]`},
	{`Include(myblock)`,
		`[{"tag":"include","attr":{"name":"myblock"}}]`},
	{`If(true) {OK}.Else {false} If(false, FALSE).ElseIf(1) {Else OK
			}.Else {Fourth}If(0).Else{ALL right}.What`,
		`[{"tag":"if","attr":{"condition":"true"},"children":[{"tag":"text","text":"OK"}],"tail":[{"tag":"else","children":[{"tag":"text","text":"false"}]}]},{"tag":"if","attr":{"condition":"false"},"children":[{"tag":"text","text":"FALSE"}],"tail":[{"tag":"elseif","attr":{"condition":"1"},"children":[{"tag":"text","text":"Else OK"}]},{"tag":"else","children":[{"tag":"text","text":"Fourth"}]}]},{"tag":"if","attr":{"condition":"0"},"tail":[{"tag":"else","children":[{"tag":"text","text":"ALL right"}]}]},{"tag":"text","text":".What"}]`},
}
