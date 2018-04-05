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

	"github.com/stretchr/testify/assert"

	"github.com/GenesisKernel/go-genesis/packages/converter"
)

func TestLimit(t *testing.T) {
	assert.NoError(t, keyLogin(1))

	rnd := randName(``)
	form := url.Values{"Name": {"tbl" + rnd}, "Columns": {`[{"name":"name","type":"number",   "conditions":"true"},
	{"name":"block", "type":"varchar","conditions":"true"}]`},
		"Permissions": {`{"insert": "true", "update" : "true", "new_column": "true"}`}}
	assert.NoError(t, postTx(`NewTable`, &form))

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
	assert.NoError(t, postTx(`NewContract`, &form))

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
	assert.NoError(t, postTx(`NewContract`, &form))

	all := 10
	sendList := func() {
		for i := 0; i < all; i++ {
			assert.NoError(t, postTx(`Limit`+rnd, &url.Values{
				`Num`:    {converter.IntToStr(i)},
				`nowait`: {`true`},
			}))
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
	assert.NoError(t, checkList(10, 1))

	var syspar ecosystemParamsResult
	assert.NoError(t, sendGet(`systemparams?names=max_tx_count,max_block_user_tx`, nil, &syspar))

	var maxusers, maxtx string
	if syspar.List[0].Name == "max_tx_count" {
		maxusers = syspar.List[1].Value
		maxtx = syspar.List[0].Value
	} else {
		maxusers = syspar.List[0].Value
		maxtx = syspar.List[1].Value
	}
	restoreMax := func() {
		assert.NoError(t, postTx(`Upd`+rnd, &url.Values{`Name`: {`max_tx_count`}, `Value`: {maxtx}}))
		assert.NoError(t, postTx(`Upd`+rnd, &url.Values{`Name`: {`max_block_user_tx`}, `Value`: {maxusers}}))
	}
	defer restoreMax()

	assert.NoError(t, postTx(`Upd`+rnd, &url.Values{`Name`: {`max_tx_count`}, `Value`: {`7`}}))

	sendList()
	assert.NoError(t, checkList(20, 3))
	assert.NoError(t, postTx(`Upd`+rnd, &url.Values{`Name`: {`max_block_user_tx`}, `Value`: {`3`}}))

	sendList()
	assert.NoError(t, checkList(30, 7))

	restoreMax()
	assert.NoError(t, sendGet(`systemparams?names=max_block_generation_time`, nil, &syspar))

	var maxtime string
	maxtime = syspar.List[0].Value
	defer func() {
		assert.NoError(t, postTx(`Upd`+rnd, &url.Values{
			`Name`:  {`max_block_generation_time`},
			`Value`: {maxtime},
		}))
	}()
	assert.NoError(t, postTx(`Upd`+rnd, &url.Values{`Name`: {`max_block_generation_time`}, `Value`: {`100`}}))

	sendList()
	assert.NoError(t, checkList(40, 0))
}
