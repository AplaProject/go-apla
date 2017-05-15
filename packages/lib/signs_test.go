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

package lib

import (
	"encoding/hex"
	"math/rand"
	"testing"

	"github.com/EGaaS/go-egaas-mvp/packages/test"
)

func TestECDSA(t *testing.T) {
	for i := 0; i < 50; i++ {
		size := rand.Intn(2049)
		if size == 0 {
			continue
		}
		forSign, _ := test.RandBytes(size)
		priv, pub, err := GenBytesKeys()
		if err != nil {
			t.Errorf(err.Error())
		}
		sign, err := SignECDSA(hex.EncodeToString(priv), string(forSign))
		if err != nil {
			t.Errorf(err.Error())
		}
		ret, err := CheckECDSA(pub, string(forSign), sign)
		if err != nil {
			t.Errorf(err.Error())
		}
		if !ret {
			t.Errorf(`ECDSA priv=%x forSign=%x sign=%x`, priv, forSign, sign)
		}
	}
}
