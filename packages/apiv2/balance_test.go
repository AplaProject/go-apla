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

package apiv2

import (
	"fmt"
	"testing"
)

func TestBalance(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	var ret balanceResult
	err := sendGet(`balance/`+gAddress, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if len(ret.Amount) < 10 {
		t.Error(`too low balance`, ret)
	}
	err = sendGet(`balance/3434341`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if len(ret.Amount) > 0 {
		t.Error(fmt.Errorf(`wrong balance %s`, ret.Amount))
		return
	}
}
