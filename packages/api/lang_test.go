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

func TestLang(t *testing.T) {

	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	name := randName(`lang`)
	value := `{"en": "English", "ru": "Русский"}`
	form := url.Values{"name": {name}, "trans": {value}}
	if err := postTx(`lang`, &form); err != nil {
		t.Error(err)
		return
	}
	ret, err := sendGet(`lang/`+name, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if ret[`trans`].(string) != value {
		t.Error(fmt.Errorf(`Lang is not right %s`, ret[`trans`].(string)))
		return
	}
	value = `{"en": "English", "it": "Italiano", "ru": "Русский"}`
	form = url.Values{"trans": {value}}

	if err := putTx(`lang/`+name, &form); err != nil {
		t.Error(err)
		return
	}
	ret, err = sendGet(`lang/`+name, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if ret[`trans`].(string) != value {
		t.Error(fmt.Errorf(`Lang is not right %s`, ret[`trans`].(string)))
		return
	}

}

func TestLangList(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	ret, err := sendGet(`langlist?limit=-1`, nil)
	if err != nil {
		t.Error(err)
		return
	}
	count := converter.StrToInt64(ret[`count`].(string))
	if count == 0 {
		t.Error(fmt.Errorf(`empty lang list`))
	}
	var (
		isYes bool
	)
	for _, item := range ret[`list`].([]interface{}) {
		mitem := item.(map[string]interface{})
		if mitem[`name`].(string) == `Yes` {
			isYes = true
			break
		}
	}
	if !isYes {
		t.Error(fmt.Errorf(`there is not yes lang`))
	}
}
