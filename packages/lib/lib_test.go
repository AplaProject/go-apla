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
	"encoding/hex"
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/test"
)

type ByteTest struct {
	src  []byte
	want []byte
}

type EncodeType struct {
	value int64
	data  []byte
}

type AddressTest struct {
	src  []byte
	want string
}

var testList = []EncodeType{
	{0, []byte{0}},
	{1, []byte{1, 1}},
	{127, []byte{1, 0x7f}},
	{65000, []byte{2, 0xe8, 0xfd}},
	{156507890, []byte{4, 0xf2, 0x1e, 0x54, 0x09}},
	{1565073467890890, []byte{7, 0xca, 0xdc, 0x19, 0x10, 0x6d, 0x8f, 0x05}},
}

func TestEncodeLenInt64(t *testing.T) {
	var off int
	buf := make([]byte, 0)
	for _, val := range testList {
		off = len(buf)
		EncodeLenInt64(&buf, val.value)
		if bytes.Compare(buf[off:len(buf)], val.data) != 0 {
			t.Errorf("different slice %d", val.value)
		}
	}
}

func TestDecodeLenInt64(t *testing.T) {
	for _, val := range testList {
		buf := val.data
		x, err := DecodeLenInt64(&buf)
		if err != nil {
			t.Error(err.Error)
		}
		if x != val.value {
			t.Errorf("different int64 %d != %d", x, val.value)
		}
	}
}

func TestAddress(t *testing.T) {
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

func TestCRCAddress(t *testing.T) {
	input := []AddressTest{
		{[]byte{23, 45, 67, 126, 64, 255, 0, 2, 1, 128}, `0035-2003-0310-5692-6089`},
		{[]byte{123, 245, 167, 26, 164, 55, 10, 12, 12, 120}, `0816-0005-5015-4925-6715`},
		{[]byte{23, 45, 67, 126, 64, 255, 0, 2, 1, 129}, `1409-9962-9733-1591-2620`},
	}
	for i, item := range input {
		address := KeyToAddress(item.src)
		if address != item.want {
			fmt.Println(`Address`, item.src, item.want)
			t.Errorf("wrong address %s != %s", item.want, address)
		}
		if !IsValidAddress(address) {
			t.Errorf("not valid address %s != %s", item.want, address)
		}
		mod := []byte(address)
		mod[len(address)-1-i] = '0' + byte(((mod[len(address)-1-i]-'0')+1)%10)
		//		fmt.Println(address, string(mod))
		if IsValidAddress(string(mod)) {
			t.Errorf("not crc %s != %s", string(mod), address)
		}
	}
}

func TestEncodeDecodeLength(t *testing.T) {
	vals := []int64{1, 67, 127, 128, 256, 1024, 2000, 10000, 65000, 1000000, 0xffeeffff,
		8123498762, 25000060000, 400000000035, -10000000044546, -1}
	for _, i := range vals {
		result := EncodeLength(i)
		got, _ := DecodeLength(&result)
		if got != i {
			t.Errorf("wrong length encoding %d != %d", i, got)
		}
	}
	if length, _ := DecodeLength(&[]byte{}); length != 0 {
		t.Errorf("wrong decoding empty slice")
	}

}

func TestFill(t *testing.T) {
	for i := 0; i < 50; i++ {
		size := rand.Intn(33)
		input, _ := test.RandBytes(size)

		out := FillLeft(input)
		if bytes.Compare(out[:32-size], make([]byte, 32-size)) != 0 ||
			bytes.Compare(out[32-size:], input) != 0 {
			t.Errorf(`different slices %x %x`, input, out)
		}
	}
}

/*
func TestEncodeBinary(t *testing.T) {
	var (
		out []byte
		off int
	)
	check := func( format string, cmp []byte, args ...interface{}) {
		if err := EncodeBinary(&out, format, args...); err!=nil {
			t.Errorf(err.Error())
		} else if bytes.Compare(out[off:], cmp) != 0 {
			t.Errorf(`different output binary data %x`, out )
		}
		off = len(out)
	}
	check( `1`, []byte{255}, 255)
	check( `414`, []byte{0,0,0x01,0x01, 0x7e, 0,1,0x86,0xa1}, 257, 126, 100001 )
	check( `ii4i`, []byte{0x01,0x43, 0x3,0x9a,0x31,1, 0,0,0xff,0xff, 0x3,0x2c,0xdd,0x15},
	               67, 78234, 0xffff, int64(1432876))
	check( `s1s`, test.HexToBytes(`0474657374c8057b0001ff86`), `test`, 200, []byte{ 123, 0, 1, 255, 134})
}*/

func TestBinMarshal(t *testing.T) {
	var out, tx []byte
	var err error
	host := `Unicode текст`
	now := Time32()
	node := test.HexToBytes(`20304350647f8f96a8`)
	_, err = BinMarshal(&out, &consts.BlockHeader{Type: 0, BlockId: 1, Time: now, WalletId: 1})
	_, err = BinMarshal(&tx, &consts.FirstBlock{TxHeader: consts.TxHeader{Type: 1, Time: now, WalletId: 1, CitizenId: 0},
		PublicKey:     test.HexToBytes(`0102300040fffa6789`),
		NodePublicKey: node,
		Host:          host})
	EncodeLenByte(&out, tx)

	tmp := hex.EncodeToString(UintToBytes(now))
	cmp := test.HexToBytes(`0000000001` + tmp + `010100002f01` + tmp +
		`010100090102300040fffa67890920304350647f8f96a812556e69636f646520d182d0b5d0bad181d182`)
	if bytes.Compare(out, cmp) != 0 {
		t.Errorf(`different output binary data %x %x`, out, cmp)
	}
	var block consts.BlockHeader
	if err = BinUnmarshal(&out, &block); err != nil {
		t.Errorf(err.Error())
	}

	//	fmt.Println( block )
	var first consts.FirstBlock
	DecodeLength(&out)
	dup := out[:]
	if err = BinUnmarshal(&out, &first); err != nil {
		t.Errorf(err.Error())
	}
	if first.Time != now || first.Host != host || first.WalletId != 1 ||
		bytes.Compare(first.NodePublicKey, node) != 0 {
		t.Errorf(`different unmarshaled %v`, first)
	}
	if len(out) != 0 {
		t.Errorf(`unfinished`)
	}
	var inter interface{}
	inter = consts.MakeStruct(`FirstBlock`)
	err = BinUnmarshal(&dup, inter)
	p := inter.(*consts.FirstBlock)
	if p.Time != now || p.Host != host || p.WalletId != 1 ||
		bytes.Compare(p.NodePublicKey, node) != 0 {
		t.Errorf(`different unmarshaled %v`, p)
	}
}

func TestHeader(t *testing.T) {
	var tx []byte
	now := Time32()
	sign := test.HexToBytes(`0056575879`)
	_, err := BinMarshal(&tx, &consts.CitizenRequest{TxHeader: consts.TxHeader{Type: 45, Time: now,
		WalletId: 123, CitizenId: 456}, StateId: 78, Sign: sign})
	if err != nil {
		t.Errorf(err.Error())
	}
	var inter interface{}
	inter = consts.MakeStruct(`CitizenRequest`)
	if err = BinUnmarshal(&tx, inter); err != nil {
		t.Errorf(err.Error())
	}
	p := inter.(*consts.CitizenRequest)
	header := reflect.ValueOf(p).Elem().Field(0)
	data := FieldToBytes(header.Interface(), 2)
	if bytes.Compare(data, test.HexToBytes(`313233`)) != 0 {
		t.Errorf(`wrong wallet_id`)
	}
	data = FieldToBytes(header.Interface(), 3)
	if bytes.Compare(data, test.HexToBytes(`343536`)) != 0 {
		t.Errorf(`wrong citizen_id`)
	}
	if bytes.Compare(consts.Sign(p), sign) != 0 {
		t.Errorf(`wrong sign`)
	}
}

func TestFieldToBytes(t *testing.T) {
	first := consts.FirstBlock{TxHeader: consts.TxHeader{Type: 1, Time: 2345,
		WalletId: 67, CitizenId: 89},
		PublicKey:     []byte(`010203`),
		NodePublicKey: []byte(`040506`),
		Host:          `070809`}
	out := ``
	for i := 1; i < 7; i++ {
		out += string(FieldToBytes(first, i))
	}
	if out != `010203040506070809` {
		t.Errorf(`different out %s`, out)
	}
	fmt.Println(out)
}
