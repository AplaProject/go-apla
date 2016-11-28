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

/*
func TestEval(t *testing.T) {
	test := []TestComp{
		{"789 63", "63"},
		{"+421", "stack is empty [1:1]"},
		{"1256778+223445", "1480223"},
		{"(67-34789)*3", "-104166"},
		{"(5+78)*(1563-527)", "85988"},
		{"124 * (143-527", "there is not pair [1:7]"},
		{"341 * 234/0", "divided by zero [1:10]"},
		{"((15+82)*2 + 5)/2", "99"},
	}
	for _, item := range test {
		out := Eval(item.Input, nil)
		if fmt.Sprint(out) != item.Output {
			t.Error(`error of eval ` + item.Input)
		}
		//		fmt.Println(out)
	}
}

func MyTable(table, id_column string, id int64, ret_column string) (int64, error) {
	if ret_column != `wallet` {
		return 0, fmt.Errorf(`Invalid result column name %s`, ret_column)
	}
	fmt.Println(table, id_column, id, ret_column)
	return 125, nil
}

func Multi(a, b int64) (int64, error) {
	return a + b*2, nil
}

func TestEvalVar(t *testing.T) {
	test := []TestComp{
		{"Multi( (34+35)*2, Multi( citizenId, 56))== 1 || Multi( (34+35)*2, Multi( citizenId, 56))== 0", `56972`},
		{"2+ Multi( (34+35)*2, Multi( citizenId, 56)) /2", `56972`},
		{"#my[id=3345].wa", "Invalid result column name wa [1:14]"},
		{"7665 + #my[id=345].wallet*2 == 7915", "true"},
		{"7665 > (citizenId-48000)", "false"},
		{"56788 + 1 >= citizenId", "true"},
		{"76 < citizenId", "true"},
		{"56789 <= citizenId", "true"},
		{"56 == 56", "true"},
		{"37 != 37", "false"},
		{"!!(1-1)", "false"},
		{"!!citizenId || wallet_id", "true"},
		{"!789", "false"},
		{"789 63", "63"},
		{"356 * ( citizenId - 50001)", "2416528"},
		{"( citizenId + wallet_id) / 2", "475120"},
		{"3* citizen_id + 2", "unknown identifier citizen_id [1:15]"},
		{"citizenId && 0", "false"},
		{"0||citizenId", "true"},
	}
	vars := map[string]interface{}{
		`citizenId`: 56789,
		`wallet_id`: 893451,
		`Multi`:     Multi,
		`Table`:     MyTable,
	}
	for _, item := range test {
		out := Eval(item.Input, &vars)
		if fmt.Sprint(out) != item.Output {
			t.Error(`error of eval ` + item.Input)
		}
		//		fmt.Println(out)
	}
}*/

func Multi(a, b int64) (int64, error) {
	return a + b*2, nil
}

func TestEvalIf(t *testing.T) {
	test := []TestComp{
		{"Multi( (34+35)*2, Multi( $citizenId, 56))== 1 || Multi( (34+35)*2, Multi( $citizenId, 56))== 0", `false`},
		{"5 + 9 > 10", `true`},
		{"34 == 45", `false`},
		{"1345", `true`},
		{"13/13-1", `false`},
		{"$citizenId == 56780 + 9", `true`},
		{"qwerty(45)", `unknown identifier qwerty`},
		/*{"Multi(2, 5) > 36", "false"},*/
	}
	vars := map[string]interface{}{
		`citizenId`: 56789,
		`wallet_id`: 893451,
		//		`Table`:     MyTable,
	}
	vm := NewVM()
	vm.Extend(&ExtendData{map[string]interface{}{"Multi": Multi}, nil})
	for i := 0; i < 2; i++ {
		for _, item := range test {
			out, err := vm.EvalIf(item.Input, &vars)
			if err != nil {
				if err.Error() != item.Output {
					t.Error(`error of ifeval ` + item.Input + err.Error())
				}
			} else {
				if fmt.Sprint(out) != item.Output {
					t.Error(`error of ifeval ` + item.Input)
				}
			}
		}
	}
}
