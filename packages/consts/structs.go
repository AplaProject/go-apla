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

package consts

import (
	"reflect"
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
	Type  byte
	Time  uint32
	KeyID int64
}

// FirstBlock is the header of FirstBlock transaction
type FirstBlock struct {
	TxHeader
	PublicKey             []byte
	NodePublicKey         []byte
	StopNetworkCertBundle []byte
	Test                  int64
	PrivateBlockchain     uint64
}

type StopNetwork struct {
	TxHeader
	StopNetworkCert []byte
}

// Don't forget to insert the structure in init() - list

var blockStructs = make(map[string]reflect.Type)

func init() {
	// New structures must be inserted here
	list := []interface{}{
		FirstBlock{},
		StopNetwork{},
	}

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
func IsStruct(tx int64) bool {
	return tx == TxTypeFirstBlock || tx == TxTypeStopNetwork
}

// Header returns TxHeader
func Header(v interface{}) TxHeader {
	return reflect.ValueOf(v).Elem().Field(0).Interface().(TxHeader)
}

// Sign returns the signature attached to the header
func Sign(v interface{}) (sign []byte) {
	field := reflect.ValueOf(v).Elem().FieldByName(`Sign`)
	if field.IsValid() && field.Kind() == reflect.Slice {
		sign = field.Bytes()
	}
	return
}
