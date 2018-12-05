// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package api

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/AplaProject/go-apla/packages/converter"
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
		   DBInsert("tbl` + rnd + `", {name: $Num, block: $block}) 
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
	assert.NoError(t, sendGet(`systemparams?names=max_tx_block,max_tx_block_per_user`, nil, &syspar))

	var maxusers, maxtx string
	if syspar.List[0].Name == "max_tx_block" {
		maxusers = syspar.List[1].Value
		maxtx = syspar.List[0].Value
	} else {
		maxusers = syspar.List[0].Value
		maxtx = syspar.List[1].Value
	}
	restoreMax := func() {
		assert.NoError(t, postTx(`Upd`+rnd, &url.Values{`Name`: {`max_tx_block`}, `Value`: {maxtx}}))
		assert.NoError(t, postTx(`Upd`+rnd, &url.Values{`Name`: {`max_tx_block_per_user`}, `Value`: {maxusers}}))
	}
	defer restoreMax()

	assert.NoError(t, postTx(`Upd`+rnd, &url.Values{`Name`: {`max_tx_block`}, `Value`: {`7`}}))

	sendList()
	assert.NoError(t, checkList(20, 3))
	assert.NoError(t, postTx(`Upd`+rnd, &url.Values{`Name`: {`max_tx_block_per_user`}, `Value`: {`3`}}))

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
