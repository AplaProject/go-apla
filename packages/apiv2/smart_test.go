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

/*
import (
	"fmt"
	"net/url"
	"strings"
	"testing"
)*/

type smartParams struct {
	Params  map[string]string
	Results map[string]string
}

type smartContract struct {
	Name   string
	Value  string
	Params []smartParams
}

/*
func TestSmartFields(t *testing.T) {

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	ret, err := sendGet(`smartcontract/MainCondition`, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if len(ret[`fields`].([]interface{})) != 0 {
		t.Error(`MainCondition fields must be empty`)
		return
	}
	if ret[`name`].(string) != `@1MainCondition` {
		t.Error(fmt.Sprintf(`MainCondition name is wrong: %s`, ret[`name`].(string)))
		return
	}
	if err := postTx(`smartcontract/MainCondition`, &url.Values{}); err != nil {
		t.Error(err)
		return
	}
}

func TestMoneyTransfer(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}

	form := url.Values{`Amount`: {`53330000`}, `Recipient`: {`3330000`}}
	if err := postTx(`smartcontract/MoneyTransfer`, &form); err != nil {
		t.Error(err)
		return
	}
}
func TestSmartContracts(t *testing.T) {

	wanted := func(name, want string) bool {
		ret, err := sendGet(`test/`+name, nil)
		if err != nil {
			t.Error(err)
			return false
		}
		if ret[`value`].(string) != want {
			t.Error(fmt.Errorf(`%s != %s`, ret[`value`].(string), want))
			return false
		}
		return true
	}

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	for _, item := range contracts {
		_, err := sendGet(`contract/`+item.Name, nil)
		if err != nil {
			if strings.Contains(err.Error(), `incorrect id`) {
				form := url.Values{"name": {item.Name}, "value": {item.Value},
					"conditions": {`true`}, `global`: {`0`}}
				if err := postTx(`contract`, &form); err != nil {
					t.Error(err)
					return
				}
				if err := putTx(`activatecontract/`+item.Name, &url.Values{}); err != nil {
					t.Error(err)
					return
				}
			} else {
				t.Error(err)
				return
			}
		}
		for _, par := range item.Params {
			form := url.Values{}
			for key, value := range par.Params {
				form[key] = []string{value}
			}
			if err := postTx(`smartcontract/`+item.Name, &form); err != nil {
				t.Error(err)
				return
			}
			for key, value := range par.Results {
				if !wanted(key, value) {
					return
				}
			}
		}
	}
}

var contracts = []smartContract{
	{`testEmpty`, `contract testEmpty {
		action { Test("empty",  "empty value")}}`,
		[]smartParams{
			{nil, map[string]string{`empty`: `empty value`}},
		}},
	{`testSimple`, `contract testSimple {
		data {
			amount int
			name   string
		}
		conditions {
			Test("scond", $amount, $name)
		}
		action { Test("sact", $name, $amount)}}`,
		[]smartParams{
			{map[string]string{`name`: `Simple name`, `amount`: `-56781`},
				map[string]string{`scond`: `-56781Simple name`,
					`sact`: `Simple name-56781`}},
		}},
}
*/
