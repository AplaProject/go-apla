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

	for _, item := range forTest {
		templ := Template2JSON(item.input, false)
		if string(templ) != item.want {
			t.Errorf(`wrong json %s != %s`, templ, item.want)
			return
		}
	}
}

var forTest = tplList{
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
	{`Button(My Contract,, myclass, NewEcosystem, "Name=myid,Id=i10,Value")`,
		`[{"tag":"button","attr":{"class":"myclass","contract":"NewEcosystem","params":{"Id":"i10","Name":"myid","Value":"Value"}},"children":[{"tag":"text","text":"My Contract"}]}]`},
	{`Div(myclass)Div()
				Div()`,
		`[{"tag":"div","attr":{"class":"myclass"}},{"tag":"div"},{"tag":"div"}]`},
	{`Div(myclass){Div()
		P(){
			Div(id){
				Label(My text,myl,forname)
			}
		}
	}`,
		`[{"tag":"div","attr":{"class":"myclass"},"children":[{"tag":"div"},{"tag":"p","children":[{"tag":"div","attr":{"class":"id"},"children":[{"tag":"label","attr":{"class":"myl","for":"forname"},"children":[{"tag":"text","text":"My text"}]}]}]}]}]`},
	{`If(true,OK)If(false){Skip}.Else{Span(Else OK)}`,
		`[{"tag":"text","text":"OK"},{"tag":"span","children":[{"tag":"text","text":"Else OK"}]}]`},
	{`If(false,First).ElseIf(0){Skip}.ElseIf(1){
		Second Span(If(text){item})
	}.ElseIf(true){Third}.Else{Fourth}`,
		`[{"tag":"text","text":"Second "},{"tag":"span","children":[{"tag":"text","text":"item"}]}]`},
	{`Button(Page: link){My Button}.Alert(ConfirmButton: ConfBtn, CancelButton: CancelBtn, 
		   Text: Alert text, Icon:myicon)`,
		`[{"tag":"button","attr":{"alert":{"cancelbutton":"CancelBtn","confirmbutton":"ConfBtn","icon":"myicon","text":"Alert text"},"page":"link"},"children":[{"tag":"text","text":"My Button"}]}]`},
}

func TestFullJSON(t *testing.T) {

	for _, item := range forFullTest {
		templ := Template2JSON(item.input, true)
		if string(templ) != item.want {
			t.Errorf(`wrong json %s != %s`, templ, item.want)
			return
		}
	}
}

var forFullTest = tplList{
	{`Simple text +=<b>bold</b>`, `[{"tag":"text","text":"Simple text +=\u0026lt;b\u0026gt;bold\u0026lt;/b\u0026gt;"}]`},
	{`Div(myclass control, Content of the Div)`, `[{"tag":"div","attr":{"class":"myclass control"},"children":[{"tag":"text","text":"Content of the Div"}]}]`},
	{`If(true,OK)If(false){Skip}.Else{Span(Else OK)}`,
		`[{"tag":"if","attr":{"condition":"true"},"children":[{"tag":"text","text":"OK"}]},{"tag":"if","attr":{"condition":"false"},"children":[{"tag":"text","text":"Skip"}],"tail":[{"tag":"else","children":[{"tag":"span","children":[{"tag":"text","text":"Else OK"}]}]}]}]`},
	{`If(false,First).ElseIf(0){Skip}.ElseIf(1){
		Second
	}.ElseIf(true){Third}.Else{Fourth}`,
		`[{"tag":"if","attr":{"condition":"false"},"children":[{"tag":"text","text":"First"}],"tail":[{"tag":"elseif","attr":{"condition":"0"},"children":[{"tag":"text","text":"Skip"}]},{"tag":"elseif","attr":{"condition":"1"},"children":[{"tag":"text","text":"Second"}]},{"tag":"elseif","attr":{"condition":"true"},"children":[{"tag":"text","text":"Third"}]},{"tag":"else","children":[{"tag":"text","text":"Fourth"}]}]}]`},
	{`Button(Page: link){My Button}.Alert(ConfirmButton: ConfBtn, CancelButton: CancelBtn, 
			Text: Alert text, Icon:myicon)`,
		`[{"tag":"button","attr":{"page":"link"},"children":[{"tag":"text","text":"My Button"}],"tail":[{"tag":"alert","attr":{"cancelbutton":"CancelBtn","confirmbutton":"ConfBtn","icon":"myicon","text":"Alert text"}}]}]`},
}
