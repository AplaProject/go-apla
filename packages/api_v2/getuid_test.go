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

package api_v2

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"testing"

	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
)

func TestGetUID(t *testing.T) {
	var ret getUIDResult
	err := sendGet(`getuid`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	gAuth = ret.Token
	/*	if ret.State == 0 {
		var instRes installResult
		err := sendPost(`install`, &url.Values{`port`: {`5432`}, `host`: {`3330000`}}, &instRes)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Println(`INSTALL`, instRes)
	}*/
	if len(ret.UID) == 0 {
		t.Errorf(`getuid has returned empty uid`)
	} else {
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
		fmt.Println(gAuth, "\r\nRefresh\r\n", lret.Refresh)
	}
}
