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

func TestContract(t *testing.T) {

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	name := randName(`cnt`)
	for _, glob := range []global{{``, `0`}, {`?global=1`, `1`}} {
		value := fmt.Sprintf(`contract %s {
			conditions {}
			action {}
	}`, name)

		form := url.Values{"name": {name}, "value": {value},
			"conditions": {`true`}, `global`: {glob.value}}
		if err := postTx(`contract`, &form); err != nil {
			t.Error(err)
			return
		}
		ret, err := sendGet(`contract/`+name+glob.url, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if ret[`value`].(string) != value {
			t.Error(fmt.Errorf(`Contract is wrong %s`, ret[`value`].(string)))
			return
		}
		value = fmt.Sprintf(`contract %s {
						conditions {
							Println("Test")
						}
						action {}
				}`, name)
		form = url.Values{"value": {value}, "conditions": {`true`}, `global`: {glob.value}}

		if err := putTx(`contract/`+ret[`id`].(string), &form); err != nil {
			t.Error(err)
			return
		}

		if err := putTx(`activatecontract/`+name+glob.url, &url.Values{}); err != nil {
			t.Error(err)
			return
		}

		ret, err = sendGet(`contract/`+ret[`id`].(string)+glob.url, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if ret[`value`].(string) != value {
			t.Error(fmt.Errorf(`Contract is wrong %s`, ret[`value`].(string)))
			return
		}

		if converter.StrToInt64(ret[`active`].(string)) == 0 {
			t.Error(fmt.Errorf(`Contract is not active`))
			return
		}
	}
}

func TestContractList(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	for _, glob := range []string{``, `?limit=10&global=1`} {
		ret, err := sendGet(`contractlist`+glob, nil)
		if err != nil {
			t.Error(err)
			return
		}
		count := converter.StrToInt64(ret[`count`].(string))
		if len(glob) == 0 {
			if count == 0 {
				t.Error(fmt.Errorf(`empty contract list`))
			}
		} else {
			if count == 0 || len(ret[`list`].([]interface{})) == 0 || len(ret[`list`].([]interface{})) > 10 {
				t.Error(fmt.Errorf(`wrong global contract list`))
			}
		}
	}
}
