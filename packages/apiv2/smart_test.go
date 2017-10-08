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
	"net/url"
	//	"strings"
	"testing"
)

type smartParams struct {
	Params  map[string]string
	Results map[string]string
}

type smartContract struct {
	Name   string
	Value  string
	Params []smartParams
}

func TestSmartFields(t *testing.T) {

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	var cntResult getContractResult
	err := sendGet(`contract/MainCondition`, nil, &cntResult)
	if err != nil {
		t.Error(err)
		return
	}
	if len(cntResult.Fields) != 0 {
		t.Error(`MainCondition fields must be empty`)
		return
	}
	if cntResult.Name != `@1MainCondition` {
		t.Error(fmt.Sprintf(`MainCondition name is wrong: %s`, cntResult.Name))
		return
	}
	if err := postTx(`MainCondition`, &url.Values{}); err != nil {
		t.Error(err)
		return
	}
}

func TestMoneyTransfer(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}

	form := url.Values{`Amount`: {`53330000`}, `Recipient`: {`0005-2070-2000-0006-0200`}}
	if err := postTx(`MoneyTransfer`, &form); err != nil {
		t.Error(err)
		return
	}
	form = url.Values{`Amount`: {`2440000`}, `Recipient`: {`1109-7770-3360-6764-7059`}, `Comment`: {`Test`}}
	if err := postTx(`MoneyTransfer`, &form); err != nil {
		t.Error(err)
		return
	}
	form = url.Values{`Amount`: {`53330000`}, `Recipient`: {`0005207000`}}
	if err := postTx(`MoneyTransfer`, &form); err != nil {
		t.Error(err)
		return
	}
}

func TestPage(t *testing.T) {

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	name := randName(`page`)
	menu := `government`
	value := `P(test,test paragraph)`

	form := url.Values{"Name": {name}, "Value": {`Param Value`},
		"Conditions": {`ContractConditions("MainCondition")`}}
	id, msg, err := postTxResult(`NewParameter`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"Name": {name}, "Value": {`New Param Value`},
		"Conditions": {`ContractConditions("MainCondition")`}}
	id, msg, err = postTxResult(`EditParameter`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"Name": {name}, "Value": {value},
		"Menu": {menu}, "Conditions": {`ContractConditions("MainCondition")`}}
	id, msg, err = postTxResult(`NewPage`, &form)
	if err != nil {
		t.Error(err)
		return
	}

	form = url.Values{"Id": {`1`}, "Value": {value + `Span(Test)`},
		"Menu": {menu}, "Conditions": {`ContractConditions("MainCondition")`}}
	id, msg, err = postTxResult(`EditPage`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(`RET`, id, msg)
	/*		ret, err := sendGet(`page/`+name+glob.url, nil)
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
		append := "P(appended, Append paragraph)"
		form = url.Values{"value": {append}, `global`: {glob.value}}

		if err := putTx(`appendpage/`+name, &form); err != nil {
			t.Error(err)
			return
		}
		ret, err = sendGet(`page/`+name+glob.url, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if ret[`value`].(string) != value+"\r\n"+append {
			t.Error(fmt.Errorf(`Appended page is not right %s`, ret[`value`].(string)))
			return
		}
	}


		wanted := func(name, want string) bool {
			var ret getTestResult
			err := sendPost(`test/`+name, nil, &ret)
			if err != nil {
				t.Error(err)
				return false
			}
			if ret.Value != want {
				t.Error(fmt.Errorf(`%s != %s`, ret.Value, want))
				return false
			}
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
				if strings.Contains(err.Error(), fmt.Sprintf(errors[`E_CONTRACT`], item.Name)) {
					form := url.Values{"Name": {item.Name}, "Value": {item.Value},
						"Conditions": {`true`}}
					if err := postTx(`NewContract`, &form); err != nil {
						if item.Name != `errTest` || !strings.HasPrefix(err.Error(), `must be type 7d01 125 [Ln:4 Col:22]`) {
							t.Error(err)
							return
						}
					}
				} else {
					t.Error(err)
					return
				}
			}
			if strings.HasSuffix(item.Name, `testUpd`) {
				continue
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
		}*/
}

/*func TestSmartContracts(t *testing.T) {

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
