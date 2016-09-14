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

func TestEvalVar(t *testing.T) {
	test := []TestComp{
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
	}
	for _, item := range test {
		out := Eval(item.Input, &vars)
		if fmt.Sprint(out) != item.Output {
			t.Error(`error of eval ` + item.Input)
		}
		fmt.Println(out)
	}
}
