// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package api

import (
	"fmt"
	"testing"

	"github.com/AplaProject/go-apla/packages/converter"
)

func TestList(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	var ret listResult
	err := sendGet(`list/contracts`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if converter.StrToInt64(ret.Count) < 7 {
		t.Error(fmt.Errorf(`The number of records %s < 7`, ret.Count))
		return
	}
	err = sendGet(`list/qwert`, nil, &ret)
	if err.Error() != `404 {"error":"E_TABLENOTFOUND","msg":"Table 1_qwert has not been found"}` {
		t.Error(err)
		return
	}
	var retTable tableResult
	for _, item := range []string{`app_params`, `parameters`} {
		err = sendGet(`table/`+item, nil, &retTable)
		if err != nil {
			t.Error(err)
			return
		}
		if retTable.Name != item {
			t.Errorf(`wrong table name %s != %s`, retTable.Name, item)
			return
		}
	}
	var sec listResult
	err = sendGet(`sections`, nil, &sec)
	if err != nil {
		t.Error(err)
		return
	}
	if converter.StrToInt(sec.Count) == 0 {
		t.Errorf(`section error`)
		return
	}
}
