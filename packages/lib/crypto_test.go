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
	"crypto/aes"
	"math/rand"
	"testing"

	"github.com/EGaaS/go-egaas-mvp/packages/test"
)

func TestPKCS7(t *testing.T) {
	for i := 0; i < 100; i++ {
		blockSize := 1 + rand.Intn(256)
		size := 1 + rand.Intn(1024)
		src, _ := test.RandBytes(size)
		pad := PKCS7Padding(src, blockSize)
		if len(pad)%blockSize != 0 || !bytes.Equal(PKCS7UnPadding(pad), src) {
			t.Errorf("PKCS7 %x != %x blocksize: %d", src, PKCS7UnPadding(pad), blockSize)
		}
	}
}

func TestCBCEncrypt(t *testing.T) {
	for i := 0; i < 50; i++ {
		key, _ := test.RandBytes(aes.BlockSize * (1 + rand.Intn(2)))
		size := 1 + rand.Intn(1024)
		src, _ := test.RandBytes(size)
		enc, err := CBCEncrypt(key, src, nil)
		if err != nil {
			t.Errorf(err.Error())
		}
		dec, err := CBCDecrypt(key, enc, nil)
		if err != nil {
			t.Errorf(err.Error())
		} else if !bytes.Equal(dec, src) {
			t.Errorf("CBCEncrypt %x != %x key: %x", src, dec, key)
		}
	}
}

func TestShared(t *testing.T) {
	for i := 0; i < 50; i++ {
		priv1, pub1, err := GenHexKeys()
		if err != nil {
			t.Errorf(err.Error())
		}
		priv2, pub2, err := GenHexKeys()
		if err != nil {
			t.Errorf(err.Error())
		}
		shared1, err := GetSharedHex(priv1, pub2)
		if err != nil {
			t.Errorf(err.Error())
		}
		shared2, err := GetSharedHex(priv2, pub1)
		if err != nil {
			t.Errorf(err.Error())
		} else if shared1 != shared2 {
			t.Errorf("Shared key %s != %s priv %x %x", shared1, shared2, priv1, priv2)
		}
	}
}

func TestSharedEncrypt(t *testing.T) {
	for i := 0; i < 50; i++ {
		size := 1 + rand.Intn(2048)
		src, _ := test.RandBytes(size)
		priv, pub, err := GenBytesKeys()
		if err != nil {
			t.Errorf(err.Error())
		}
		enc, err := SharedEncrypt(pub, src)
		if err != nil {
			t.Errorf(err.Error())
		}
		dec, err := SharedDecrypt(priv, enc)
		if err != nil {
			t.Errorf(err.Error())
		} else if !bytes.Equal(dec, src) {
			t.Errorf("SharedEncrypt %x != %x priv key: %x", src, dec, priv)
		}
	}
}
