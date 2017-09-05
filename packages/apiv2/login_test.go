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
	/*	"encoding/hex"
		"fmt"
		"net/url"*/
	"testing"
	//	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
)

func TestAuth(t *testing.T) {
	/*	ret, err := sendGet(`getuid`, nil)
		if err != nil {
			t.Error(err)
			return
		}
		if uid, ok := ret[`uid`]; !ok {
			t.Errorf(`getuid has returned empty uid`)
		} else {
			priv, pub, err := crypto.GenHexKeys()
			if err != nil {
				t.Error(err)
				return
			}
			sign, err := crypto.Sign(priv, uid.(string))
			if err != nil {
				t.Error(err)
				return
			}
			form := url.Values{"pubkey": {pub}, "signature": {hex.EncodeToString(sign)}}
			ret, err = sendPost(`login`, &form)
			if err != nil {
				t.Error(err)
				return
			}
		}*/
}

/*func TestBadRequest(t *testing.T) {
	form := url.Values{"pubkey": {`001122`}}
	ret, err := sendPost(`login`, &form)
	if ret != nil {
		t.Error(fmt.Errorf(`must be 400 error`))
		return
	}
	if err.Error() != `400 Value signature is undefined` {
		t.Error(fmt.Errorf(`wrong error: "%s"`, err.Error()))
		return
	}
}

func TestPageNotFound(t *testing.T) {
	ret, err := sendGet(`test`, nil)
	if ret != nil {
		t.Error(fmt.Errorf(`must be 404 error`))
		return
	}
	if err.Error() != `404 404 page not found` {
		t.Error(fmt.Errorf(`wrong error: "%s"`, err.Error()))
		return
	}
}
*/
