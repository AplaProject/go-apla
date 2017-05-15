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
	"bytes"
	"math/rand"
	"testing"

	"github.com/EGaaS/go-egaas-mvp/packages/test"
)

func TestStringAddress(t *testing.T) {
	for i := 0; i < 50; i++ {
		key, _ := test.RandBytes(64)
		address := Address(key)
		wallet := AddressToString(address)
		if StringToAddress(wallet) != address {
			t.Errorf("StringToAddress %d for %s Key: %x", address, wallet, key)
		}
	}
}
func TestKeyToAddress(t *testing.T) {
	for i := 0; i < 50; i++ {
		key, seed := test.RandBytes(64)
		address := KeyToAddress(key)
		if (i % 10) == 0 {
			if IsValidAddress(address[:len(address)-1]) {
				t.Errorf("valid address %s for %x seed: %d", address[:len(address)-1], key, seed)
			}
		} else if !IsValidAddress(address) {
			t.Errorf("not valid address %s for %x seed: %d", address, key, seed)
		}
	}
}
func TestFill(t *testing.T) {
	for i := 0; i < 50; i++ {
		size := rand.Intn(33)
		input, _ := test.RandBytes(size)

		out := FillLeft(input)
		if !bytes.Equal(out[:32-size], make([]byte, 32-size)) || !bytes.Equal(out[32-size:], input) {
			t.Errorf(`different slices %x %x`, input, out)
		}
	}
}

func TestGenKeys(t *testing.T) {
	for i := 0; i < 50; i++ {
		priv, pub, err := GenHexKeys()
		if err != nil {
			t.Errorf(err.Error())
		}
		getpub := PrivateToPublicHex(priv)
		if len(getpub) != 128 || getpub != pub {
			t.Errorf(`different pubkeys %s %s for private %s`, pub, getpub, priv)
		}
	}
}

type checkTest struct {
	src  string
	want int
}

func TestCheckSum(t *testing.T) {
	var data = []checkTest{
		{`0123`, 6},
		{`23785238`, 0},
		{`178902332005238`, 9},
		{`0943735134343343438`, 0},
		{`-ashdgediwewe2369+[]`, 3},
	}
	for i, item := range data {
		if item.want != CheckSum([]byte(item.src)) {
			t.Errorf(`different CheckSum %s num: %d`, item.src, i)
		}
	}
}
