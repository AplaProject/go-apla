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

package template

import (
	"testing"
)

type tplItem struct {
	input string
	want  string
}

type tplList []tplItem

func TestJSON(t *testing.T) {
	var timeout bool
	vars := make(map[string]string)
	vars[`_full`] = `0`
	for _, item := range forTest {
		templ := Template2JSON(item.input, &timeout, &vars)
		if string(templ) != item.want {
			t.Errorf("wrong json \r\n%s != \r\n%s", templ, item.want)
			return
		}
	}
}

var forTest = tplList{
	{`Button(Body: addpage, 
		Contract: NewPage, 
		Params: "Name=hello_page2, Value=Div(fefe, dbbt), Menu=default_menu, Conditions=true")`,
		`[{"tag":"button","attr":{"contract":"NewPage","params":{"Conditions":{"text":"true","type":"text"},"Menu":{"text":"default_menu","type":"text"},"Name":{"text":"hello_page2","type":"text"},"Value":{"params":["fefe","dbbt"],"type":"Div"}}},"children":[{"tag":"text","text":"addpage"}]}]`},
	{"Button(Body: add table1, Contract: NewTable, Params: `Name=name,Columns=[{\"name\":\"MyName\",\"type\":\"varchar\", \"index\": \"1\",  \"conditions\":\"true\"}, {\"name\":\"Amount\", \"type\":\"number\",\"index\": \"0\", \"conditions\":\"true\"}],Permissions={\"insert\": \"true\", \"update\" : \"true\", \"new_column\": \"true\"}`)", `[{"tag":"button","attr":{"contract":"NewTable","params":{"Columns":{"text":"[{\"name\":\"MyName\",\"type\":\"varchar\", \"index\": \"1\",  \"conditions\":\"true\"}, {\"name\":\"Amount\", \"type\":\"number\",\"index\": \"0\", \"conditions\":\"true\"}]","type":"text"},"Name":{"text":"name","type":"text"},"Permissions":{"text":"{\"insert\": \"true\", \"update\" : \"true\", \"new_column\": \"true\"}","type":"text"}}},"children":[{"tag":"text","text":"add table1"}]}]`},
	{`Calculate( Exp: 342278783438/0, Type: money )Calculate( Exp: 5.2/0, Type: float )
		Calculate( Exp: 7/0)`,
		`[{"tag":"text","text":"dividing by zerodividing by zerodividing by zero"}]`},
	{`SetVar(val, 2200000034343443343430000)SetVar(zero, 0)Calculate( Exp: (342278783438+5000)*(#val#-932780000), Type: money, Prec:18 )Calculate( Exp: (2+50)*(#zero#-9), Type: money )`,
		`[{"tag":"text","text":"753013346318631859.1075080680647-468"}]`},
	{`SetVar(val, 100)Calculate(10000-(34+5)*#val#)=Calculate("((10+#val#-45)*3.0-10)/4.5 + #val#", Prec: 4)`,
		`[{"tag":"text","text":"6100"},{"tag":"text","text":"=141.1111"}]`},
	{`Span((span text), ok )Span(((span text), ok) )Div(){{My body}}`,
		`[{"tag":"span","attr":{"class":"ok"},"children":[{"tag":"text","text":"(span text)"}]},{"tag":"span","children":[{"tag":"text","text":"((span text), ok)"}]},{"tag":"div","children":[{"tag":"text","text":"{My body}"}]}]`},
	{`Code(P(Some text)
 Div(myclass){
	 Span(Strong("Bold text"))
 })`,
		`[{"tag":"code","attr":{"text":"P(Some text)\n Div(myclass){\n\t Span(Strong(\"Bold text\"))\n }"}}]`},
	{`Data(Source: mysrc, Columns: "id,name", Data:
		1, First Name
		2, Second Name
	).Custom(custom_id){
		SetVar(Name: v, Value: Lower(#name#))
		P(Body: #v#)
	}.Custom(cust){
		P(Body: #v#)
	}Data(Columns: "name", Data:
		First Name
		Second Name
	)`,
		`[{"tag":"data","attr":{"columns":["id","name","custom_id","cust"],"data":[["1","First Name","[{\"tag\":\"p\",\"children\":[{\"tag\":\"text\",\"text\":\"first name\"}]}]","[{\"tag\":\"p\",\"children\":[{\"tag\":\"text\",\"text\":\"first name\"}]}]"],["2","Second Name","[{\"tag\":\"p\",\"children\":[{\"tag\":\"text\",\"text\":\"second name\"}]}]","[{\"tag\":\"p\",\"children\":[{\"tag\":\"text\",\"text\":\"second name\"}]}]"]],"source":"mysrc","types":["text","text","tags","tags"]}},{"tag":"data","attr":{"columns":["name"],"data":[["First Name"],["Second Name"]],"types":["text"]}}]`},

	{`Data(Source: mysrc, Columns: "id,name", Data:
		1,first
		2,second
		3,third
	).Custom("synthetic"){
		Div(text-muted, #name#)
	}
	Table(Source: mysrc)`, `[{"tag":"data","attr":{"columns":["id","name","synthetic"],"data":[["1","first","[{\"tag\":\"div\",\"attr\":{\"class\":\"text-muted\"},\"children\":[{\"tag\":\"text\",\"text\":\"first\"}]}]"],["2","second","[{\"tag\":\"div\",\"attr\":{\"class\":\"text-muted\"},\"children\":[{\"tag\":\"text\",\"text\":\"second\"}]}]"],["3","third","[{\"tag\":\"div\",\"attr\":{\"class\":\"text-muted\"},\"children\":[{\"tag\":\"text\",\"text\":\"third\"}]}]"]],"source":"mysrc","types":["text","text","tags"]}},{"tag":"table","attr":{"source":"mysrc"}}]`},
	{`Data(myforlist,"id,name",
		"1",Test message 1
		2,"Test message 2"
		3,"Test message 3"
		)ForList(nolist){Problem}ForList(myforlist){
			Div(){#id#. Em(#name#)}
		}`,
		`[{"tag":"data","attr":{"columns":["id","name"],"data":[["1","Test message 1"],["2","Test message 2"],["3","Test message 3"]],"source":"myforlist","types":["text","text"]}},{"tag":"forlist","attr":{"source":"myforlist"},"children":[{"tag":"div","children":[{"tag":"text","text":"1. "},{"tag":"em","children":[{"tag":"text","text":"Test message 1"}]}]},{"tag":"div","children":[{"tag":"text","text":"2. "},{"tag":"em","children":[{"tag":"text","text":"Test message 2"}]}]},{"tag":"div","children":[{"tag":"text","text":"3. "},{"tag":"em","children":[{"tag":"text","text":"Test message 3"}]}]}]}]`},
	{`SetTitle(My pageР)AddToolButton(Title: Open, Page: default)`,
		`[{"tag":"settitle","attr":{"title":"My pageР"}},{"tag":"addtoolbutton","attr":{"page":"default","title":"Open"}}]`},
	{`DateTime(2017-11-07T17:51:08)+DateTime(2015-08-27T09:01:00,HH:MI DD.MM.YYYY)
	+CmpTime(2017-11-07T17:51:08,2017-11-07)CmpTime(2017-11-07T17:51:08,2017-11-07T20:22:01)CmpTime(2015-10-01T17:51:08,2015-10-01T17:51:08)=DateTime(NULL)`,
		`[{"tag":"text","text":"2017-11-07 17:51:08"},{"tag":"text","text":"+09:01 27.08.2015"},{"tag":"text","text":"\n\t+1-10"},{"tag":"text","text":"="}]`},
	{`SetVar(pref,unicode Р)Input(Name: myid, Value: #pref#)Strong(qqq)`,
		`[{"tag":"input","attr":{"name":"myid","value":"unicode Р"}},{"tag":"strong","children":[{"tag":"text","text":"qqq"}]}]`},
	{`ImageInput(myimg,100,40)`,
		`[{"tag":"imageinput","attr":{"name":"myimg","ratio":"40","width":"100"}}]`},
	{`LinkPage(My page,mypage,,"myvar1=Value 1, myvar2=Value2,myvar3=Val(myval)")`,
		`[{"tag":"linkpage","attr":{"page":"mypage","pageparams":{"myvar1":{"text":"Value 1","type":"text"},"myvar2":{"text":"Value2","type":"text"},"myvar3":{"params":["myval"],"type":"Val"}}},"children":[{"tag":"text","text":"My page"}]}]`},
	{`Image(/images/myimage.jpg,My photo,myclass).Style(width:100px;)`,
		`[{"tag":"image","attr":{"alt":"My photo","class":"myclass","src":"/images/myimage.jpg","style":"width:100px;"}}]`},
	{`Data(mysrc,"id,name",
		"1",John Silver,2
		2,"Mark, Smith"
	)`,
		`[{"tag":"data","attr":{"columns":["id","name"],"data":[],"error":"line 2, column 0: wrong number of fields in line","source":"mysrc","types":["text","text"]}}]`},
	{`Select(myselect,mysrc,name,id,0,myclass)`,
		`[{"tag":"select","attr":{"class":"myclass","name":"myselect","namecolumn":"name","source":"mysrc","value":"0","valuecolumn":"id"}}]`},
	{`Data(mysrc,"id,name"){
		"1",John Silver
		2,"Mark, Smith"
		3,"Unknown ""Person"""
		}`,
		`[{"tag":"data","attr":{"columns":["id","name"],"data":[["1","John Silver"],["2","Mark, Smith"],["3","Unknown \"Person\""]],"source":"mysrc","types":["text","text"]}}]`},
	{`If(true) {OK}.Else {false} Div(){test} If(false, FALSE).ElseIf(0) { Skip }.ElseIf(1) {Else OK
		}.Else {Fourth}If(0).Else{ALL right}`,
		`[{"tag":"text","text":"OK"},{"tag":"div","children":[{"tag":"text","text":"test"}]},{"tag":"text","text":"Else OK"},{"tag":"text","text":"ALL right"}]`},
	{`Button(Contract: MyContract, Body:My Contract, Class: myclass, Params:"Name=myid,Id=i10,Value")`,
		`[{"tag":"button","attr":{"class":"myclass","contract":"MyContract","params":{"Id":{"text":"i10","type":"text"},"Name":{"text":"myid","type":"text"},"Value":{"text":"Value","type":"text"}}},"children":[{"tag":"text","text":"My Contract"}]}]`},
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
		`[{"tag":"button","attr":{"class":"myclass","contract":"NewEcosystem","params":{"Id":{"text":"i10","type":"text"},"Name":{"text":"myid","type":"text"},"Value":{"text":"Value","type":"text"}},"style":".btn {\n\t\tborder: 10px 10px;\n\t}"},"children":[{"tag":"text","text":"My Contract"}]}]`},
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
		`[{"tag":"menuitem","attr":{"page":"page1","title":"Menu 1"}},{"tag":"menugroup","attr":{"name":"SubMenu","title":"SubMenu"},"children":[{"tag":"menuitem","attr":{"page":"page2","title":"Menu 2"}},{"tag":"menuitem","attr":{"icon":"person","page":"page3","title":"Menu 3"}}]}]`},
	{`SetVar(testvalue, The, #n#, Value).(n, New).(param,"23")Span(Test value equals #testvalue#).(#param#)`,
		`[{"tag":"span","children":[{"tag":"text","text":"Test value equals The, New, Value"}]},{"tag":"span","children":[{"tag":"text","text":"23"}]}]`},
	{`SetVar(test, mytest).(empty,0)And(0,test,0)Or(0,#test#)Or(0, And(0,0))And(0,Or(0,my,while))
		And(1,#mytest#)Or(#empty#, And(#empty#, line))`,
		`[{"tag":"text","text":"010010"}]`},
	{`Address()Span(Address(-5728238900021))Address(3467347643873).(-6258391547979339691)`,
		`[{"tag":"text","text":"unknown address"},{"tag":"span","children":[{"tag":"text","text":"1844-6738-3454-7065-1595"}]},{"tag":"text","text":"0000-0003-4673-4764-38731218-8352-5257-3021-1925"}]`},
	{`Table(src, "ID=id,name,Wallet=wallet")`,
		`[{"tag":"table","attr":{"columns":[{"Name":"id","Title":"ID"},{"Name":"name","Title":"name"},{"Name":"wallet","Title":"Wallet"}],"source":"src"}}]`},
	{`Chart(Type: "bar", Source: src, FieldLabel: "name", FieldValue: "count", Colors: "red, green")`,
		`[{"tag":"chart","attr":{"colors":["red","green"],"fieldlabel":"name","fieldvalue":"count","source":"src","type":"bar"}}]`},
}

func TestFullJSON(t *testing.T) {
	var timeout bool
	vars := make(map[string]string)
	vars[`_full`] = `1`
	for _, item := range forFullTest {
		templ := Template2JSON(item.input, &timeout, &vars)
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
