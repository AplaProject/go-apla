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

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
)

func TestMenu(t *testing.T) {

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	name := randName(`menu`)
	ret, err := sendGet(`menu/government`, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if len(ret[`value`].(string)) == 0 {
		t.Error(fmt.Errorf(`empty get menu`))
		return
	}
	for _, glob := range []global{{`?global=0`, `0`}, {`?global=1`, `1`}} {
		value := `MenuItem(Test,test)`

		form := url.Values{"name": {name}, "value": {value},
			"conditions": {`true`}, `global`: {glob.value}}
		if err := postTx(`menu`, &form); err != nil {
			t.Error(err)
			return
		}
		ret, err = sendGet(`menu/`+name+glob.url, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if ret[`value`].(string) != value {
			t.Error(fmt.Errorf(`Menu is not right %s`, ret[`value`].(string)))
			return
		}

		value += "\r\nMenuItem(Updated, updated)"
		form = url.Values{"value": {value},
			"conditions": {`true`}, `global`: {glob.value}}

		if err := putTx(`menu/`+name, &form); err != nil {
			t.Error(err)
			return
		}
		ret, err = sendGet(`menu/`+name+glob.url, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if ret[`value`].(string) != value {
			t.Error(fmt.Errorf(`Menu is not right %s`, ret[`value`].(string)))
			return
		}

		append := "MenuItem(Appended, appended)"
		form = url.Values{"value": {append}, `global`: {glob.value}}

		if err := putTx(`appendmenu/`+name, &form); err != nil {
			t.Error(err)
			return
		}
		ret, err = sendGet(`menu/`+name+glob.url, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if ret[`value`].(string) != value+"\r\n"+append {
			t.Error(fmt.Errorf(`Appended menu is not right %s`, ret[`value`].(string)))
			return
		}
	}
}

func TestMenuList(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	for _, glob := range []string{``, `?limit=2&global=1`} {
		ret, err := sendGet(`menulist`+glob, nil)
		if err != nil {
			t.Error(err)
			return
		}
		count := converter.StrToInt64(ret[`count`].(string))
		if len(glob) == 0 {
			if count == 0 {
				t.Error(fmt.Errorf(`empty menu list`))
			}
			var (
				isGov, isDef bool
			)
			for _, item := range ret[`list`].([]interface{}) {
				mitem := item.(map[string]interface{})
				if mitem[`name`].(string) == `government` {
					isGov = true
				} else if mitem[`name`].(string) == `menu_default` {
					isDef = true
				}
			}
			if !isGov {
				t.Error(fmt.Errorf(`there is not government menu`))
			}
			if !isDef {
				t.Error(fmt.Errorf(`there is not default menu`))
			}
		} else {
			if count == 0 || len(ret[`list`].([]interface{})) == 0 || len(ret[`list`].([]interface{})) > 2 {
				t.Error(fmt.Errorf(`wrong global menu list`))
			}
		}
	}
}
