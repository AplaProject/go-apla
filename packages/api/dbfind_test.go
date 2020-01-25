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
	"net/url"
	"testing"
)

func TestDbFind(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	var ret dbFindResult

	// Query table that is known to be existing
	if err := sendPost("dbfind/keys", &url.Values{}, &ret); nil != err {
		t.Error(err)
		return
	}
	if length := len(ret.List); 0 == length {
		t.Error(fmt.Errorf(`The number of records %d = 0`, length))
		return
	}

	// Query table that doesn't exist
	if err := sendPost("dbfind/QA_not_existing_table", &url.Values{}, &ret); nil != err {
		if err.Error() != `404 {"error":"E_TABLENOTFOUND","msg":"Table qa_not_existing_table has not been found"}` {
			t.Error(err)
			return
		}
	}

	// Query table with specified columns
	if err := sendPost("dbfind/keys", &url.Values{"Columns": {"id, account"}}, &ret); nil != err {
		t.Error(err)
	}

	// Query table with specified columns that are known to be missing
	if err := sendPost("dbfind/keys", &url.Values{"Columns": {"id,account,id_non_existing"}}, &ret); nil != err {
		if err.Error() != `400 {"error":"E_SERVER","msg":"column id_non_existing doesn't exist"}` {
			t.Error(err)
			return
		}
	}
}
