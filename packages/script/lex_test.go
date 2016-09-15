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

package script

import (
	"fmt"
	"testing"
)

type TestLexem struct {
	Input  string
	Output string
}

func (lexems Lexems) String(source []rune) (ret string) {
	for _, item := range lexems {
		slex := string(source[item.Offset:item.Right])
		if item.Type == 0 {
			slex = `error`
		}
		ret += fmt.Sprintf("[%d %s]", item.Type, slex)
	}
	return
}

func TestLexParser(t *testing.T) {
	test := []TestLexem{
		{`(ab <= 24 )|| (12>67) && (56==78)`, `[1 (][4 ab][2 <=][3 24][1 )][2 ||][1 (][3 12][2 >][3 67][1 )][2 &&][1 (][3 56][2 ==][3 78][1 )]`},
		{`!ab < !b && 12>=56 && qwe!=asd`, `[2 !][4 ab][2 <][2 !][4 b][2 &&][3 12][2 >=][3 56][2 &&][4 qwe][2 !=][4 asd]`},
		{`ab || 12 && 56`, `[4 ab][2 ||][3 12][2 &&][3 56]`},
		{`true | 42`, `[4 true][0 error]`},
		{"(\r\n)\x03 -", "[1 (][1 \n][1 )][0 error]"},
		{` +( - )	/ `, `[2 +][1 (][2 -][1 )][2 /]`},
		{`23+13424 Тест`, `[3 23][2 +][3 13424][4 Тест]`},
		{` 0785/67+iname*(56-31)`, `[3 0785][2 /][3 67][2 +][4 iname][2 *][1 (][3 56][2 -][3 31][1 )]`},
		{`myvar_45 - a_qwe + t81you - 345rt`, `[4 myvar_45][2 -][4 a_qwe][2 +][4 t81you][2 -][0 error]`},
		{`10 + #mytable[id = 234].name * 20`, `[3 10][2 +][1 #][4 mytable][1 [][4 id][2 =][3 234][1 ]][1 .][4 name][2 *][3 20]`},
	}
	for _, item := range test {
		source := []rune(item.Input)
		out := LexParser(source)
		if out.String(source) != item.Output {
			t.Error(`error of lexical parser ` + item.Input)
		}
		//		fmt.Println(out.String(source))
	}
}
