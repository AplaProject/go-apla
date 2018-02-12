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
	"time"

	"github.com/GenesisKernel/go-genesis/packages/converter"
)

func TestLimit(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	rnd := randName(``)
	form := url.Values{"Name": {"tbl" + rnd}, "Columns": {`[{"name":"name","type":"number",   "conditions":"true"},
	{"name":"block", "type":"varchar","conditions":"true"}]`},
		"Permissions": {`{"insert": "true", "update" : "true", "new_column": "true"}`}}
	err := postTx(`NewTable`, &form)
	if err != nil {
		t.Error(err)
		return
	}
	form = url.Values{`Value`: {`contract Limit` + rnd + ` {
		data {
			Num int
		}
		conditions {
		}
		action {
		   DBInsert("tbl` + rnd + `", "name, block", $Num, $block) 
		}
	}`}, `Conditions`: {`true`}}
	if err := postTx(`NewContract`, &form); err != nil {
		t.Error(err)
		return
	}

	form = url.Values{`Value`: {`contract Upd` + rnd + ` {
		data {
			Name string
			Value string
		}
		conditions {
		}
		action {
		   DBUpdateSysParam($Name, $Value, "") 
		}
	}`}, `Conditions`: {`true`}}
	if err := postTx(`NewContract`, &form); err != nil {
		t.Error(err)
		return
	}

	all := 10
	sendList := func() {
		for i := 0; i < all; i++ {
			if err := postTx(`Limit`+rnd, &url.Values{`Num`: {converter.IntToStr(i)}, `nowait`: {`true`}}); err != nil {
				t.Error(err)
				return
			}
		}
		time.Sleep(10 * time.Second)
	}
	checkList := func(count, wantBlocks int) (err error) {
		var list listResult
		err = sendGet(`list/tbl`+rnd, nil, &list)
		if err != nil {
			return
		}
		if converter.StrToInt(list.Count) != count {
			return fmt.Errorf(`wrong list items %s != %d`, list.Count, count)
		}
		blocks := make(map[string]int)
		for _, item := range list.List {
			if v, ok := blocks[item["block"]]; ok {
				blocks[item["block"]] = v + 1
			} else {
				blocks[item["block"]] = 1
			}
		}
		if wantBlocks > 0 && len(blocks) != wantBlocks {
			return fmt.Errorf(`wrong number of blocks %d != %d`, len(blocks), wantBlocks)
		}
		return nil
	}
	sendList()
	if err = checkList(10, 1); err != nil {
		t.Error(err)
		return
	}
	var syspar ecosystemParamsResult
	err = sendGet(`systemparams?names=max_tx_count,max_block_user_tx`, nil, &syspar)
	if err != nil {
		t.Error(err)
		return
	}
	var maxusers, maxtx string

	if syspar.List[0].Name == "max_tx_count" {
		maxusers = syspar.List[1].Value
		maxtx = syspar.List[0].Value
	} else {
		maxusers = syspar.List[0].Value
		maxtx = syspar.List[1].Value
	}
	restoreMax := func() {
		if err := postTx(`Upd`+rnd, &url.Values{`Name`: {`max_tx_count`}, `Value`: {maxtx}}); err != nil {
			t.Error(err)
			return
		}
		if err := postTx(`Upd`+rnd, &url.Values{`Name`: {`max_block_user_tx`}, `Value`: {maxusers}}); err != nil {
			t.Error(err)
			return
		}
	}
	defer restoreMax()
	if err := postTx(`Upd`+rnd, &url.Values{`Name`: {`max_tx_count`}, `Value`: {`7`}}); err != nil {
		t.Error(err)
		return
	}
	sendList()
	if err = checkList(20, 3); err != nil {
		t.Error(err)
		return
	}
	if err := postTx(`Upd`+rnd, &url.Values{`Name`: {`max_block_user_tx`}, `Value`: {`3`}}); err != nil {
		t.Error(err)
		return
	}
	sendList()
	if err = checkList(30, 7); err != nil {
		t.Error(err)
		return
	}
	restoreMax()
	err = sendGet(`systemparams?names=max_block_generation_time`, nil, &syspar)
	if err != nil {
		t.Error(err)
		return
	}
	var maxtime string
	maxtime = syspar.List[0].Value
	defer func() {
		if err := postTx(`Upd`+rnd, &url.Values{`Name`: {`max_block_generation_time`},
			`Value`: {maxtime}}); err != nil {
			t.Error(err)
			return
		}
	}()
	if err := postTx(`Upd`+rnd, &url.Values{`Name`: {`max_block_generation_time`}, `Value`: {`100`}}); err != nil {
		t.Error(err)
		return
	}
	sendList()
	if err = checkList(40, 0); err != nil {
		t.Error(err)
		return
	}

}
