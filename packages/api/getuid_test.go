// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package api

import (
	"encoding/hex"
	"encoding/json"
	"net/url"
	"testing"

	"github.com/GenesisKernel/go-genesis/packages/crypto"
)

func TestGetUID(t *testing.T) {
	var ret getUIDResult
	err := sendGet(`getuid`, nil, &ret)
	if err != nil {
		var v map[string]string
		json.Unmarshal([]byte(err.Error()[4:]), &v)
		if v[`error`] == `E_NOTINSTALLED` {
			var instRes installResult
			err := sendPost(`install`, &url.Values{`db_port`: {`5432`}, `db_host`: {`localhost`},
				`type`: {`PRIVATE_NET`}, `db_name`: {`apla`}, `log_level`: {`ERROR`},
				`db_pass`: {`postgres`}, `db_user`: {`postgres`}}, &instRes)
			if err != nil {
				t.Error(err)
			}
			return
		} else {
			t.Error(err)
			return
		}
	}
	gAuth = ret.Token
	priv, pub, err := crypto.GenHexKeys()
	if err != nil {
		t.Error(err)
		return
	}
	sign, err := crypto.Sign(priv, ret.UID)
	if err != nil {
		t.Error(err)
		return
	}
	form := url.Values{"pubkey": {pub}, "signature": {hex.EncodeToString(sign)}}
	var lret loginResult
	err = sendPost(`login`, &form, &lret)
	if err != nil {
		t.Error(err)
		return
	}
	gAuth = lret.Token
	var ref refreshResult
	err = sendPost(`refresh`, &url.Values{"token": {lret.Refresh}}, &ref)
	if err != nil {
		t.Error(err)
		return
	}
	gAuth = ref.Token
}
