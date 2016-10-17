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

package consts

import (
	"reflect"
)

type BlockHeader struct {
	Type     byte
	BlockId  uint32
	Time     uint32
	WalletId int64
	CBId     byte
	Sign     []byte
}

type TxHeader struct {
	Type      byte
	Time      uint32
	WalletId  int64
	CitizenId int64
}

type TXHeader struct {
	Type     int32 // byte < 128 system tx 129 - 1 byte 130 - 2 bytes 131 - 3 - bytes 132 - 4 bytes
	Time     uint32
	WalletId uint64
	StateId  int32
	Flags    uint8
	Sign     []byte
}

type TXNewCitizen struct {
	TXHeader
	PublicKey []byte
}

type FirstBlock struct {
	TxHeader
	PublicKey     []byte
	NodePublicKey []byte
	Host          string
}

type CitizenRequest struct {
	TxHeader
	StateId int64
	Sign    []byte
}

type NewCitizen struct {
	TxHeader
	StateId   int64
	PublicKey []byte
	Sign      []byte
}

// Don't forget to insert the structure in init() - list

var blockStructs = make(map[string]reflect.Type)

func init() {
	list := []interface{}{FirstBlock{}, CitizenRequest{}, NewCitizen{}, TXNewCitizen{}} // New structures must be inserted here

	for _, item := range list {
		blockStructs[reflect.TypeOf(item).Name()] = reflect.TypeOf(item)
	}
}

func MakeStruct(name string) interface{} {
	v := reflect.New(blockStructs[name]) //.Elem()
	return v.Interface()
}

func IsStruct(tx int) bool {
	return tx > 0 && tx <= 4 /*TXNewCitizen*/
}

func Header(v interface{}) TxHeader {
	return reflect.ValueOf(v).Elem().Field(0).Interface().(TxHeader)
}

func HeaderNew(v interface{}) TXHeader {
	return reflect.ValueOf(v).Elem().Field(0).Interface().(TXHeader)
}

func Sign(v interface{}) (sign []byte) {
	field := reflect.ValueOf(v).Elem().FieldByName(`Sign`)
	if field.IsValid() && field.Kind() == reflect.Slice {
		sign = field.Bytes()
	}
	return
}
