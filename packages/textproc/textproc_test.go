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
	//	"fmt"
	"testing"
)

type TestText struct {
	src  string
	want string
}

func TestDo(t *testing.T) {
	input := []TestText{
		{`test #string#`, `test #string#`},
		{`test par#string`, `test par#string`},
		{`#val1# строка`, `строка 1 строка`},
		{`test par#string #value2#`, `test par#string test строка 1 test`},
		{`#value2#`, `test строка 1 test`},
		{`prefix #var##val1#`, `prefix строка 1 + test строка 1 testстрока 1`},
		{`example #loop#`, `example qwer qwer qwer qwer qwer qwer qwer qwer qwer qwer qwer #loop# post post post post post post post post post post post`},
	}
	vars := map[string]string{
		`val1`: `строка 1`, `value2`: `test #val1# test`,
		`var`: `#val1# + #value2#`, `loop`: `qwer #loop# post`,
	}
	for _, item := range input {
		get := Do(item.src, &vars)
		if get != item.want {
			t.Errorf(`wrong result %s != %s`, get, item.want)
		}
	}
}
