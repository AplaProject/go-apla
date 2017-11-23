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
	"encoding/hex"
	"encoding/json"
	"net/url"
	"testing"

	"github.com/AplaProject/go-apla/packages/crypto"
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
				return
			}
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
