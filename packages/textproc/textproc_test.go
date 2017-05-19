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

package textproc

import (
	"fmt"
	"strings"
	"testing"
)

type TestText struct {
	src  string
	want string
}

var (
	vars = map[string]string{
		`val1`: `строка 1`, `value2`: `test #val1# test`,
		`var`: `#val1# + #value2#`, `loop`: `qwer #loop# post`,
	}
)

func FullName(vars *map[string]string, pars ...string) string {
	return strings.Join(pars, ` `)
}

func TestDo(t *testing.T) {
	AddFuncs(&map[string]TextFunc{`FullName`: FullName, `AsIs`: AsIs})
	AddMaps(&map[string]MapFunc{`Map1`: Map1, `Table1`: Table1})
	input := []TestText{
		{`FullName(Param, qwert) test #FullName(Test, #val1#) OK(eeee) #string# FullName(qqq, #var#)`,
			`Param qwert test Test строка 1 OK(eeee) #string# qqq строка 1 + test строка 1 test`},
		{`test #Map1{href: http://google.com, Name: "test, quote"} and #NoFunc()`, `test (http://google.com:test, quote) and #NoFunc()`},
		{`test #FullName(First Name, Last Name) and #NoFunc() and #AsIs("(finish)")`, `test First Name Last Name and #NoFunc() and (finish)`},
		{`test #string#`, `test #string#`},
		{`test par#string`, `test par#string`},
		{`#val1# строка`, `строка 1 строка`},
		{`test par#string #value2#`, `test par#string test строка 1 test`},
		{`#value2#`, `test строка 1 test`},
		{`prefix #var##val1#`, `prefix строка 1 + test строка 1 testстрока 1`},
		{`example #loop#`, `example qwer qwer qwer qwer qwer qwer qwer qwer qwer qwer qwer #loop# post post post post post post post post post post post`},
	}
	for _, item := range input {
		get := Macro(item.src, &vars)
		if get != item.want {
			t.Errorf(`wrong result %s != %s`, get, item.want)
		}
	}
}

func AsIs(vars *map[string]string, pars ...string) string {
	return pars[0]
}

func JSON(vars *map[string]string, pars ...string) string {
	fmt.Println(`JSON`, pars)
	if len(pars) == 0 {
		return ``
	}
	return fmt.Sprintf(`<script language="JavaScript" type="text/javascript">
	var jdata = { 
%s 
}
</script>`, pars[0])
}

func TestFunc(t *testing.T) {
	AddFuncs(&map[string]TextFunc{`AsIs`: AsIs, `Json`: JSON})

	input := []TestText{
		{`AsIs : span, ("строка")`, `span, ("строка")`},
		{`Link(http://google.com, "test, quote")BR()`, `<a href="http://google.com" title="">test, quote</a><br>`},
		{`Link(http://google.com, Google)`, `<a href="http://google.com" title="">Google</a>`},
		{`Link(http://#value2#, Tag(b, Site #val1#)), Title)`, `<a href="http://test строка 1 test" title=""><b>Site строка 1</b></a>`},
		{`Link(http://google.com, Google)Tag(div, Text1	Text 2)`,
			`<a href="http://google.com" title="">Google</a><div>Text1	Text 2</div>`},
	}
	for _, item := range input {
		get := Process(item.src, &vars)
		if get != item.want {
			t.Errorf(`wrong result [%s] != [%s]`, get, item.want)
		}
	}
}

func Map1(vars *map[string]string, pars *map[string]string) string {
	return fmt.Sprintf("(%s:%s)", (*pars)[`href`], (*pars)[`Name`])
}

func Table1(vars *map[string]string, pars *map[string]string) string {
	return fmt.Sprintf("Table(%s:%s)", (*pars)[`Table`], *Split((*pars)[`Column`]))
}

func TestMap(t *testing.T) {

	AddMaps(&map[string]MapFunc{`Map1`: Map1, `Table1`: Table1})
	input := []TestText{
		{`Map1{ href: http://google.com, Name: "test, quote"}
		Map1{ href: #val1#,
			Name: #value2# }`, `(http://google.com:test, quote)
(строка 1:test строка 1 test)`},
		{`Table1{ Table: #val1#_table
				Column: [[ID, #value2#], [Name, #val1# ooops], [Name, Call(#val1#, ooops) ]]}`,
			`Table(строка 1_table:[[ID  #value2#] [Name  #val1# ooops] [Name  Call(#val1#, ooops) ]])`},
	}
	for _, item := range input {
		get := Process(item.src, &vars)
		if get != item.want {
			t.Errorf(`wrong result %s != %s`, get, item.want)
		}
	}
}
