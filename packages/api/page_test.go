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
	"net/url"
	"testing"
)

func TestPage(t *testing.T) {

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	name := randName(`page`)
	menu := `government`
	for _, glob := range []global{{``, `0`}, {`/global`, `1`}} {
		value := `P(test,test paragraph)`

		form := url.Values{"name": {name}, "value": {value},
			"menu": {menu}, "conditions": {`true`}, `global`: {glob.value}}
		if err := postTx(`page`, &form); err != nil {
			t.Error(err)
			return
		}
		ret, err := sendGet(`page/`+name+glob.url, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if ret[`value`].(string) != value {
			t.Error(fmt.Errorf(`Menu is not right %s`, ret[`value`].(string)))
			return
		}

		value += "\r\nP(updated, Additional paragraph)"
		form = url.Values{"value": {value}, "menu": {menu},
			"conditions": {`true`}, `global`: {glob.value}}

		if err := putTx(`page/`+name, &form); err != nil {
			t.Error(err)
			return
		}
		ret, err = sendGet(`page/`+name+glob.url, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if ret[`value`].(string) != value {
			t.Error(fmt.Errorf(`Page is not right %s`, ret[`value`].(string)))
			return
		}
		if ret[`menu`].(string) != menu {
			t.Error(fmt.Errorf(`Page menu is not right %s`, ret[`menu`].(string)))
			return
		}
	}
}
