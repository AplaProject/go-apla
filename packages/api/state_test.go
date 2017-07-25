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
	"strings"
	"testing"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
)

func TestState(t *testing.T) {

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	name := randName(`state`)
	form := url.Values{"name": {name}, "currency": {strings.ToUpper(name)}}
	if err := postTx(`newstate`, &form); err != nil {
		t.Error(err)
		return
	}
	ret, err := sendGet(`statelist`, nil)
	if err != nil {
		t.Error(err)
		return
	}
	var id int64
	for _, item := range ret[`list`].([]interface{}) {
		mitem := item.(map[string]interface{})
		if mitem[`name`].(string) == name {
			id = converter.StrToInt64(mitem[`id`].(string))
			break
		}
	}
	if id == 0 {
		t.Error(fmt.Errorf(`unknown state id`))
		return
	}
	if err := keyLogin(id); err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"name": {`test`}, "value": {`test value`}, "conditions": {`true`}}
	if err := postTx(`stateparams`, &form); err != nil {
		t.Error(err)
		return
	}
	newval := `new test value`
	form = url.Values{"value": {newval}, "conditions": {`true`}}
	if err := putTx(`stateparams/test`, &form); err != nil {
		t.Error(err)
		return
	}

	ret, err = sendGet(`stateparams/test`, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if ret[`value`].(string) != newval {
		t.Error(fmt.Errorf(`wrong test value %s`, ret[`value`].(string)))
		return
	}
	var isCurrency bool
	ret, err = sendGet(`stateparamslist`, nil)
	if err != nil {
		t.Error(err)
		return
	}
	for _, item := range ret[`list`].([]interface{}) {
		mitem := item.(map[string]interface{})
		if mitem[`name`].(string) == `currency_name` && mitem[`value`].(string) == strings.ToUpper(name) {
			isCurrency = true
			break
		}
	}
	if !isCurrency {
		t.Error(fmt.Errorf(`wrong currency value`))
		return
	}
}
