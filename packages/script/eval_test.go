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

func Multi(a, b int64) (int64, error) {
	return a + b*2, nil
}

func TestEvalIf(t *testing.T) {
	test := []TestComp{
		{`Multi(45, $citizenId")`, `there is not pair`},
		{"34 + `45` < 0", `runtime panic error`},
		{"Multi( (34+35)*2, Multi( $citizenId, 56))== 1 || Multi( (34+35)*2, Multi( $citizenId, 56))== 0", `false`},
		{"5 + 9 > 10", `true`},
		{"34 == 45", `false`},
		{"1345", `true`},
		{"13/13-1", `false`},
		{"7665 > ($citizenId-48000)", "false"},
		{"56788 + 1 >= $citizenId", "true"},
		{"76 < $citizenId", "true"},
		{"56789 <= $citizenId", "true"},
		{"56 == 56", "true"},
		{"37 != 37", "false"},
		{"!!(1-1)", "false"},
		{"!!$citizenId || $wallet_id", "true"},
		{"!789", "false"},
		{"$citizenId == 56780 + 9", `true`},
		{"qwerty(45)", `unknown identifier qwerty`},
		{"Multi(2, 5) > 36", "false"},
		{"789 63 == 63", "true"},
		{"+421", "stack is empty"},
		{"1256778+223445==1480223", "true"},
		{"(67-34789)*3 == -104166", "true"},
		{"(5+78)*(1563-527) == 85988", "true"},
		{"124 * (143-527", "there is not pair"},
		{"341 * 234/0", "divided by zero"},
		{"0 == ((15+82)*2 + 5)/2 - 99", "true"},
		{"Multi( (34+35)*2, Multi( $citizenId, 56))== 1 || Multi( (34+35)*2, Multi( $citizenId, 56))== 0", `false`},
		{"2+ Multi( (34+35)*2, Multi( $citizenId, 56)) /2 == 56972", `true`},
		{"$citizenId && 0", "false"},
		{"0|| ($citizenId + $wallet_id == 950240)", "true"},
	}
	vars := map[string]interface{}{
		`citizenId`: 56789,
		`wallet_id`: 893451,
	}
	vm := NewVM()
	vm.Extend(&ExtendData{map[string]interface{}{"Multi": Multi}, nil, nil})
	for _, item := range test {
		out, err := vm.EvalIf(item.Input, 0, &vars)
		if err != nil {
			if err.Error() != item.Output {
				t.Error(`error of ifeval ` + item.Input + ` ` + err.Error())
			}
		} else {
			if fmt.Sprint(out) != item.Output {
				t.Error(`error of ifeval ` + item.Input + ` Output:` + fmt.Sprint(out))
			}
		}
	}
}
