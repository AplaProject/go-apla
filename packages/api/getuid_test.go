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

	"github.com/GenesisKernel/go-genesis/packages/crypto"
)

type PubSign struct {
	Pub  string
	Sign string
}

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
	sign, err := crypto.Sign(priv, nonceSalt+ret.UID)
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
	list := []PubSign{
		{`b0a7bfd6a5bbc9e30a116721e232a7718510178bb22e35a40e09d7933b1a343fa7137f916d1f7198360a6c5c47c29ad38bfe3f097a793002e99847040be00a8`,
			`3045022100c19d0e133b60de85eaa4bd1373cc940e1bab978baca06c3bc83da3b51fd5877f0220664f16d71a0e3bed39ee28dcbbc6df8efedebc2e743ad39d896987cbef3d7b2f`},
		{`3df7dcede40579ae7f818a5a4402cc3ea90fbc9b286514c76b28c5c02c6f36d23bdc3c7a282f07274a0a1a61da2921fa2f6961f846f959b5cf8e7cee570699`,
			`30450221009a6084dac666a2630775adf288279937a64caaddf6c000c1fbf4e2f50ac02f9b02206fd7bf1b9f04f3bbf02cab9f1de4e9066b3f653548175b0ff9a3ae42218b11ea`},
	}
	for _, item := range list {
		form = url.Values{"pubkey": {item.Pub}, "signature": {item.Sign}}
		err = sendPost(`login`, &form, &lret)
		if err != nil {
			t.Error(err)
		}
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
