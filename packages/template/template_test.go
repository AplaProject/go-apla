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
	"fmt"
	"testing"

	"github.com/AplaProject/go-apla/packages/config"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/textproc"
)

type tempItem struct {
	input string
	want  string
}

type tempList []tempItem

func init() {
	var (
		err error
	)
	config.ConfigIni = map[string]string{
		`db_user`:     `postgres`,
		`db_port`:     `5432`,
		`db_name`:     `test`,
		`db_password`: `postgres`,
		`log_level`:   `ERROR`,
		`db_type`:     `postgresql`,
		`db_host`:     `localhost`,
	}
	err = model.GormInit(config.ConfigIni["db_user"], config.ConfigIni["db_password"], config.ConfigIni["db_name"])
	if err != nil {
		fmt.Println(`Connect error`)
	}
}

func TestSanitize(t *testing.T) {
	params := make(map[string]string)
	params[`state_id`] = `1`
	params[`global`] = `0`

	for _, item := range inputSanitize {
		templ := textproc.Process(item.input, &params)
		if templ != item.want {
			t.Errorf(`wrong sanitize %s != %s`, templ, item.want)
			return
		}
	}
}

var inputSanitize = tempList{
	{`LangJS("myres")`, `<span class="lang" lang-id="myres"></span>`},
	{`Money(12"><script></script>.34)`, `12scriptscript.34`},
	{`If(1," onclick=""alert('false');""")`, `onclick=&#34;alert(&#39;false&#39;);&#34;`},
	{`Now(tatete"")`, `tatete`},
	{`Textarea(eatete" onclick="alert('qq'), "myclass data=weyu", <script>text<script> )`,
		`<textarea id="eatete onclickalertqq" class="myclass" data="weyu">&lt;script&gt;text&lt;script&gt;</textarea>`},
	{`Input(volume,form-control input-lg qwert=uwiw onclick=ioese", "", "number"" onclick=""alert(123)""", 0 )`,
		`<input type="number onclickalert123" id="volume" placeholder="" class="form-control input-lg" value="0" qwert="uwiw" onclick="ioese">`},
	{`InputDate(mydate,qwert=uwiw, "12-07-2017"" onclick=""alert(123)""")`,
		`<input type="text" class="datetimepicker " qwert="uwiw" id="mydate" value="12-07-2017&#34; onclick=&#34;alert(123)&#34;">`},
	{`InputMoney(mymoney "edede",uwiw, "-12.072017 ")`,
		`<input id="mymoney edede" type="text" value="-12.072017" data-inputmask="'alias': 'numeric', 'rightAlign': false, 'groupSeparator': ' ', 
'autoGroup': true, 'digits': 2, 'digitsOptional': false, 'prefix': '', 'placeholder': '0'"	class="inputmask uwiw">`},
	{`Trim("<script>""test"" ")`, `&lt;script&gt;&#34;test&#34;`},
	{`Back(template, mypage, "}]);alert('qq');")`, `<script language="JavaScript" type="text/javascript">
hist_push(['load_template', 'mypage', {)alert('qq')}]);</script>`},
	//	{`Json(template: "qqqq"}; alert('test'); var i={)`, ``},
	{`Li(Small(form,Title<script>alert('OK');</script>), my)`,
		`<li class="my" ><small class="form" >Title&lt;script&gt;alert('OK');</script&gt;</small></li>`},
}
