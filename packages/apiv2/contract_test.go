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
	//	"fmt"
	"net/url"
	"strings"
	"testing"
)

func TestSmartContracts(t *testing.T) {

	wanted := func(name, want string) bool {
		/*		var ret getContractResult
				err := sendGet(`test/`+name, nil, &ret)
				if err != nil {
					t.Error(err)
					return false
				}
				if ret[`value`].(string) != want {
					t.Error(fmt.Errorf(`%s != %s`, ret[`value`].(string), want))
					return false
				}*/
		return true
	}

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	for _, item := range contracts {
		var ret getContractResult
		err := sendGet(`contract/`+item.Name, nil, &ret)
		if err != nil {
			if strings.Contains(err.Error(), `there is not `+item.Name+` contract`) {
				form := url.Values{"Name": {item.Name}, "Value": {item.Value},
					"Conditions": {`true`}}
				if err := postTx(`NewContract`, &form); err != nil {
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
			if err := postTx(item.Name, &form); err != nil {
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
