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

package api

import (
	"fmt"
	"testing"
)

func TestContent(t *testing.T) {

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	ret, err := sendGet(`content/page/government`, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if ret[`html`].(string) == `NULL` || len(ret[`html`].(string)) == 0 {
		t.Error(fmt.Errorf(`empty page`))
		return
	}
	ret, err = sendGet(`content/page/government/global`, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if len(ret[`html`].(string)) != 0 {
		t.Error(fmt.Errorf(`not empty global page`))
		return
	}
	ret, err = sendGet(`content/menu/government`, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if ret[`html`].(string) == `NULL` || len(ret[`html`].(string)) == 0 {
		t.Error(fmt.Errorf(`empty menu`))
		return
	}
	ret, err = sendGet(`content/menu/government/global`, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if len(ret[`html`].(string)) != 0 {
		t.Error(fmt.Errorf(`not empty global menu`))
		return
	}
}
