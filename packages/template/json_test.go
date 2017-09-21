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

	for _, item := range forTest {
		templ := Template2JSON(item.input)
		if string(templ) != item.want {
			t.Errorf(`wrong json %s != %s`, templ, item.want)
			return
		}
	}
}

var forTest = tplList{
	{`LangJS("myres")`, `[{"type":"fn","name":"LangJS","data":["myres"]}]`},
	{`Money(12"><script>.34)Small(ok wedwedwe)`, `[{"type":"fn","name":"Money","data":["12\"\u003e\u003cscript\u003e.34"]},{"type":"fn","name":"Small","data":["ok wedwedwe"]}]`},
	{`P(myclass," onclick=""alert('false');""")`, `[{"type":"fn","name":"P","data":["myclass","onclick=\"alert('false');\""]}]`},
	{`Now()`, `[{"type":"fn","name":"Now"}]`},
	{`Trim("<script>""test"" ")`, `[{"type":"fn","name":"Trim","data":["\u003cscript\u003e\"test\""]}]`},
	{`Li(Small(form,Title alert('OK'))P(pclass, "My paragraph"), my)`,
		`[{"type":"fn","name":"Li","data":[[{"type":"fn","name":"Small","data":["form","Title alert('OK')"]},{"type":"fn","name":"P","data":["pclass","My paragraph"]}],"my"]}]`},
	{`P(pclass form, At the biginning Small(form,Title) at the end)`,
		`[{"type":"fn","name":"P","data":["pclass form",["At the biginning ",{"type":"fn","name":"Small","data":["form","Title"]}," at the end"]]}]`},
	{`Div("pdiv form", "At the biginning P(pcl, Small(form,Title string))")`,
		`[{"type":"fn","name":"Div","data":["pdiv form",["At the biginning ",{"type":"fn","name":"P","data":["pcl",[{"type":"fn","name":"Small","data":["form","Title string"]}]]}]]}]`},
	{`Divs(myclass1)
	  DivsEnd:`,
		`[{"type":"block","name":"Divs","data":["myclass1"],"children":[{"type":"fn","name":"DivsEnd"}]}]`},
	{`Divs:myclass1 cl2
		P(pcl, Small(form,Title string))
	  DivsEnd:
	  `, `[{"type":"block","name":"Divs","data":["myclass1 cl2"],"children":[{"type":"fn","name":"P","data":["pcl",[{"type":"fn","name":"Small","data":["form","Title string"]}]]},{"type":"fn","name":"DivsEnd"}]}]`},
	{`Divs(level1 form, level2)
		P(pcl, The first line)
		P(pcl type, The second line)
	DivsEnd:
`, `[{"type":"block","name":"Divs","data":["level1 form","level2"],"children":[{"type":"fn","name":"P","data":["pcl","The first line"]},{"type":"fn","name":"P","data":["pcl type","The second line"]},{"type":"fn","name":"DivsEnd"}]}]`},
	{`TxForm{Contract: MyContract}`, `[{"type":"map","name":"TxForm","map":{"Contract":"MyContract"}}]`},
	{`TxButton{Contract: "MyTest", 
		Inputs: "Name=myname, Request #= myreq",
		OnSuccess: "MyPage, RequestId:# myreq#"}`, `[{"type":"map","name":"TxButton","map":{"Contract":"MyTest","Inputs":"Name=myname, Request #= myreq","OnSuccess":"MyPage, RequestId:# myreq#"}}]`},
	{`AutoUpdate(5)
		Table{
		  Table:  citizens
		  Order: id
		  Columns: [[Avatar, Image(#avatar#)],  [ID, Address(#id#)],  [Name, #name#]]
		}
	  AutoUpdateEnd:`, `[{"type":"block","name":"AutoUpdate","data":["5"],"children":[{"type":"map","name":"Table","map":{"":"","Columns":"[[Avatar, Image(#avatar#)],  [ID, Address(#id#)],  [Name, #name#]]","Order":"id","Table":"citizens"}},{"type":"fn","name":"AutoUpdateEnd"}]}]`},

	{`Title: State info
	  Navigation(LiTemplate(government,Government), State info)
	  Divs(md-4, panel panel-default elastic center)
		  Divs: panel-body
			IfParams(#flag#=="", Image(static/img/noflag.svg, No flag, img-responsive),Image(#flag#, Flag, img-responsive))
		  DivsEnd:
	  DivsEnd:
`, `[{"type":"fn","name":"Title","data":["State info"]},{"type":"fn","name":"Navigation","data":[[{"type":"fn","name":"LiTemplate","data":["government","Government"]}],"State info"]},{"type":"block","name":"Divs","data":["md-4","panel panel-default elastic center"],"children":[{"type":"block","name":"Divs","data":["panel-body"],"children":[{"type":"fn","name":"IfParams","data":["#flag#==\"\"",[{"type":"fn","name":"Image","data":["static/img/noflag.svg","No flag","img-responsive"]}],[{"type":"fn","name":"Image","data":["#flag#","Flag","img-responsive"]}]]},{"type":"fn","name":"DivsEnd"}]},{"type":"fn","name":"DivsEnd"}]}]`},
}
