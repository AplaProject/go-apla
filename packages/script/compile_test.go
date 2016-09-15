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

type TestComp struct {
	Input  string
	Output string
}

func (bytecode Bytecodes) String(source []rune) (ret string) {
	for _, item := range bytecode {
		if item.Cmd == CMD_ERROR {
			item.Value = item.Value.(string) + fmt.Sprintf(` [Ln:%d Col:%d]`, item.Lex.Line, item.Lex.Column)
		}
		ret += fmt.Sprintf("[%d %v]", item.Cmd, item.Value)
	}
	return
}

func TestCompile(t *testing.T) {
	test := []TestComp{
		{`10 + #mytable[id = 234].name * 20`, `[1 10][3 mytable][3 id][1 234][3 name][4 50][1 20][514 30][512 25]`},
		{"!!12 + !!0", "[1 12][256 50][256 50][1 0][256 50][256 50][512 25]"},
		{"12346 7890", "[1 12346][1 7890]"},
		{"460+ 1540", "[1 460][1 1540][512 25]"},
		{"10 - 2 *3", "[1 10][1 2][1 3][514 30][513 25]"},
		{"20/5 + 78 * 23*1", "[1 20][1 5][515 30][1 78][1 23][514 30][1 1][514 30][512 25]"},
		{"5*(2 + 3)", "[1 5][1 2][1 3][512 25][514 30]"},
		{"(67-23)*45 + (2*7-56)/100", "[1 67][1 23][513 25][1 45][514 30][1 2][1 7][514 30][1 56][513 25][1 100][515 30][512 25]"},
		{"5*(25 / (3+2) - 1)", "[1 5][1 25][1 3][1 2][512 25][515 30][1 1][513 25][514 30]"},
		{"(8 +(3+2*((33-11))))", "[1 8][1 3][1 2][1 33][1 11][513 25][514 30][512 25][512 25]"},
		{"(8 +11))+56", "[1 8][1 11][512 25][0 there is not pair ) [Ln:1 Col:8]]"},
		{"(99+ 76)(1+67", "[1 99][1 76][512 25][1 1][1 67][512 25][0 there is not pair [Ln:1 Col:9]]"},
		{"678 || 34 && 768 + 56", "[1 678][1 34][1 768][1 56][512 25][516 15][517 10]"},
	}
	for _, item := range test {
		source := []rune(item.Input)
		out := Compile(source).String(source)

		if out != item.Output {
			t.Error(`error of compile ` + item.Input)
		}
		//		fmt.Println(out)
	}
}
