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

const (
	// TxfPublic is the flag of TXHeader
	TxfPublic = 0x01 // Append PublicKey to Sign in TXHeader
)

// BlockHeader is a structure of the block header
type BlockHeader struct {
	Type     byte
	BlockID  uint32
	Time     uint32
	WalletID int64
	StateID  byte
	Sign     []byte
}

// TxHeader is the old version of the transaction header
type TxHeader struct {
	Type      byte
	Time      uint32
	WalletID  int64
	CitizenID int64
}

// TXHeader is the header of the contract's transactions
type TXHeader struct {
	Type     int32 // byte < 128 system tx 129 - 1 byte 130 - 2 bytes 131 - 3 - bytes 132 - 4 bytes
	Time     uint32
	WalletID uint64
	StateID  int32
	Flags    uint8
	Sign     []byte
}

// TXNewCitizen isn't used
type TXNewCitizen struct {
	TXHeader
	PublicKey []byte
}

// FirstBlock is the header of FirstBlock transaction
type FirstBlock struct {
	TxHeader
	PublicKey     []byte
	NodePublicKey []byte
	Host          string
}

// CitizenRequest isn't used
type CitizenRequest struct {
	TxHeader
	StateID int64
	Sign    []byte
}

// NewCitizen isn't used
type NewCitizen struct {
	TxHeader
	StateID   int64
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

// MakeStruct is only used for FirstBlock now
func MakeStruct(name string) interface{} {
	v := reflect.New(blockStructs[name]) //.Elem()
	return v.Interface()
}

// IsStruct is only used for FirstBlock now
func IsStruct(tx int) bool {
	return tx == 1 // > 0 && tx <= 4 /*TXNewCitizen*/
}

// Header returns TxHeader
func Header(v interface{}) TxHeader {
	return reflect.ValueOf(v).Elem().Field(0).Interface().(TxHeader)
}

// HeaderNew returns TXHeader
func HeaderNew(v interface{}) TXHeader {
	return reflect.ValueOf(v).Elem().Field(0).Interface().(TXHeader)
}

// Sign returns the signature attached to the header
func Sign(v interface{}) (sign []byte) {
	field := reflect.ValueOf(v).Elem().FieldByName(`Sign`)
	if field.IsValid() && field.Kind() == reflect.Slice {
		sign = field.Bytes()
	}
	return
}
