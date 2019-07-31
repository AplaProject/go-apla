// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package api

import (
	"encoding/hex"
	"encoding/json"
	"net/url"
	"testing"

	"github.com/AplaProject/go-apla/packages/crypto"

	"github.com/stretchr/testify/assert"
)

func TestGetUID(t *testing.T) {
	var ret getUIDResult
	err := sendGet(`getuid`, nil, &ret)
	if err != nil {
		var v map[string]string
		json.Unmarshal([]byte(err.Error()[4:]), &v)
		t.Error(err)
		return
	}
	gAuth = ret.Token
	priv, pub, err := crypto.GenHexKeys()
	if err != nil {
		t.Error(err)
		return
	}
	sign, err := crypto.SignString(priv, `LOGIN`+ret.NetworkID+ret.UID)
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
}

func TestNetwork(t *testing.T) {
	var ret NetworkResult
	assert.NoError(t, sendGet(`network`, nil, &ret))
	if len(ret.NetworkID) == 0 || len(ret.CentrifugoURL) == 0 || len(ret.FullNodes) == 0 {
		t.Error(`Wrong value`, ret)
	}
}
