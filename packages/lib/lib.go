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
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"reflect"
	"time"

	b58 "github.com/jbenet/go-base58"
	"golang.org/x/crypto/ripemd160"
)

// Converts binary address to DayLight address.
func BytesToAddress(address []byte) string {
	return `D` + b58.Encode(address)
}

// DecodeLenInt64 gets int64 from []byte and shift the slice. The []byte should  be
// encoded with EncodeLengthPlusInt64.
func DecodeLenInt64(data *[]byte) (int64, error) {
	if len(*data) == 0 {
		return 0, nil
	}
	length := int((*data)[0]) + 1
	if len(*data) < length {
		return 0, fmt.Errorf(`length of data %d < %d`, len(*data), length)
	}
	buf := make([]byte, 8)
	copy(buf, (*data)[1:length])
	x := int64(binary.LittleEndian.Uint64(buf))
	*data = (*data)[length:]
	return x, nil
}

// Encodes values into binary data. The format parameter can contain the following characters:
// 1 - 1 byte for encoding byte, int8, uint8
// 4 - 4 bytes for encoding int32, uint32
// i - 2-9 bytes for encoding int64, uint64 by EncodeLenInt64 function
// s - for encoding string or []byte by EncodeLenByte function
/*func EncodeBinary(out *[]byte, format string, args ...interface{}) error {
	if *out == nil {
		*out = make([]byte, 0, 2048)
	}
	if len(format) != len(args) {
		return fmt.Errorf(`wrong count of parameters %d != %d`, len(format), len(args))
	}
	tmp := make([]byte,4)
	for i, ch := range format {
		switch ch {
			case '1', '4':
				switch ival := args[i].(type) {
					case int8, uint8, int, int32, uint32:
						val,_ := ival.(int)
						if ch == '1' {
							*out = append(*out, uint8(val))
						} else {
							binary.BigEndian.PutUint32(tmp, uint32(val))
							*out = append(*out, tmp...)
						}
					default:
						return fmt.Errorf(`wrong type %d`, i)
				}
			case 'i':
				switch ival := args[i].(type) {
					case int8, uint8, int, int32, uint32:
						val,_ := ival.(int)
						EncodeLenInt64(out, int64(val))
					case int64, uint64:
						val,_ := ival.(int64)
						EncodeLenInt64(out, val)
					default:
						return fmt.Errorf(`wrong type %d`, i)
				}
			case 's':
				switch ival := args[i].(type) {
					case string:
						EncodeLenByte(out, []byte(ival))
					case []byte:
						EncodeLenByte(out, ival)
					default:
						return fmt.Errorf(`wrong type %d`, i)
				}
			default:
				return fmt.Errorf(`unknown input binary format`)
		}
	}
	return nil
}*/

// Convert 32-byte value into [4]byte (BigEndian)
func UintToBytes(val uint32) []byte {
	tmp := make([]byte, 4)
	binary.BigEndian.PutUint32(tmp, val)
	return tmp
}

// Encodes int64 number to []byte. If it is less than 128 then it returns []byte{length}.
// Otherwise, it returns (0x80 | len of int64) + int64 as BigEndian []byte
//
//   67 => 0x43
//   1024 => 0x820400
//   1000000 => 0x830f4240
//
func EncodeLength(length int64) []byte {
	if length >= 0 && length <= 127 {
		return []byte{byte(length)}
	}
	buf := make([]byte, 9)
	binary.BigEndian.PutUint64(buf[1:], uint64(length))
	i := 1
	for ; buf[i] == 0 && i < 8; i++ {
	}
	buf[0] = 0x80 | byte(9-i)
	return append(buf[:1], buf[i:]...)
}

// Decodes []byte to int64 and shifts buf. Bytes must be encoded with EncodeLength function.
//
//   0x43 => 67
//   0x820400 => 1024
//   0x830f4240 => 1000000
//
func DecodeLength(buf *[]byte) (ret int64, err error) {
	if len(*buf) == 0 {
		return
	}
	length := (*buf)[0]
	if (length & 0x80) != 0 {
		length &= 0x7F
		if len(*buf) < int(length+1) {
			return 0, fmt.Errorf(`input slice has small size`)
		}
		ret = int64(binary.BigEndian.Uint64(append(make([]byte, 8-length), (*buf)[1:length+1]...)))
	} else {
		ret = int64(length)
		length = 0
	}
	*buf = (*buf)[length+1:]
	return
}

// Appends the length of the slice (EncodeLength) + the slice.
func EncodeLenByte(out *[]byte, buf []byte) *[]byte {
	*out = append(append(*out, EncodeLength(int64(len(buf)))...), buf...)
	return out
}

// EncodeLenInt64 appends int64 to []byte as uint8 + little-endian order of uint8.
//
//  65000 => [0x02, 0xe8, 0xfd]
//
func EncodeLenInt64(data *[]byte, x int64) *[]byte {
	var length int
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(x))
	for length = 8; length > 0 && buf[length-1] == 0; length-- {
	}
	*data = append(append(*data, byte(length)), buf[:length]...)
	return data
}

// Fill the slice by zero at left if the size of the slice is less than 32.
func FillLeft(slice []byte) []byte {
	if len(slice) >= 32 {
		return slice
	}
	return append(make([]byte, 32-len(slice)), slice...)
}

// Function generate a random pair of ECDSA private and public keys.
func GenKeys() (privKey string, pubKey string) {
	private, _ := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	privKey = hex.EncodeToString(private.D.Bytes())
	pubKey = hex.EncodeToString(append(FillLeft(private.PublicKey.X.Bytes()), FillLeft(private.PublicKey.Y.Bytes())...))
	return
}

// Function IsValidAddress checks if the specified address is DayLight address.
func IsValidAddress(address string) bool {
	if address[0] != 'D' {
		return false
	}
	key := b58.Decode(address[1:])
	checksum := key[len(key)-4:]
	finger := key[:len(key)-4]
	h256 := sha256.Sum256(finger)
	h256 = sha256.Sum256(h256[:])
	return bytes.Compare(checksum, h256[:4]) == 0
}

func Address(pubKey []byte) []byte {
	h256 := sha256.Sum256(pubKey)
	h := ripemd160.New()
	h.Write(h256[:])
	finger := h.Sum(nil)
	h256 = sha256.Sum256(finger)
	h256 = sha256.Sum256(h256[:])
	checksum := h256[:4]
	return append(finger, checksum...)
}

// Converts a public key to DayLight address.
func KeyToAddress(pubKey []byte) string {
	return BytesToAddress(Address(pubKey))
}

// Tiem gets the current time in UNIX format.
func Time32() uint32 {
	return uint32(time.Now().Unix())
}

func BinMarshal(out *[]byte, v interface{}) (*[]byte, error) {
	t := reflect.ValueOf(v)
	if *out == nil {
		*out = make([]byte, 0, 2048)
	}

	switch t.Kind() {
	case reflect.Uint8, reflect.Int8:
		*out = append(*out, uint8(t.Uint()))
	case reflect.Uint32:
		tmp := make([]byte, 4)
		binary.BigEndian.PutUint32(tmp, uint32(t.Uint()))
		*out = append(*out, tmp...)
	case reflect.Int32:
		if uint32(t.Int()) < 128 {
			*out = append(*out, uint8(t.Int()))
		} else {
			var i uint8
			tmp := make([]byte, 4)
			binary.BigEndian.PutUint32(tmp, uint32(t.Int()))
			for ; i < 4; i++ {
				if tmp[i] != uint8(0) {
					break
				}
			}
			*out = append(*out, uint8(128+4-i))
			*out = append(*out, tmp[i:]...)
		}
	case reflect.Int64, reflect.Uint64:
		EncodeLenInt64(out, t.Int())
	case reflect.String:
		*out = append(append(*out, EncodeLength(int64(t.Len()))...), []byte(t.String())...)
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			BinMarshal(out, t.Field(i).Interface())
		}
	case reflect.Slice:
		*out = append(append(*out, EncodeLength(int64(t.Len()))...), t.Bytes()...)
	case reflect.Ptr:
		BinMarshal(out, t.Elem().Interface())
	default:
		return out, fmt.Errorf(`unsupported type of BinMarshal`)
	}
	return out, nil
}

func BinUnmarshal(out *[]byte, v interface{}) error {
	t := reflect.ValueOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if len(*out) == 0 {
		return fmt.Errorf(`input slice is empty`)
	}
	switch t.Kind() {
	case reflect.Uint8, reflect.Int8:
		val := uint64((*out)[0])
		t.SetUint(val)
		*out = (*out)[1:]
	case reflect.Uint32:
		t.SetUint(uint64(binary.BigEndian.Uint32((*out)[:4])))
		*out = (*out)[4:]
	case reflect.Int32:
		val := (*out)[0]
		if val < 128 {
			t.SetInt(int64(val))
			*out = (*out)[1:]
		} else {
			var i uint8
			size := val - 128
			tmp := make([]byte, 4)
			for ; i < size; i++ {
				tmp[4-size+i] = (*out)[i+1]
			}
			t.SetInt(int64(binary.BigEndian.Uint32(tmp)))
			*out = (*out)[size+1:]
		}
	case reflect.Int64, reflect.Uint64:
		if val, err := DecodeLenInt64(out); err != nil {
			return err
		} else {
			t.SetInt(val)
		}
	case reflect.String:
		if val, err := DecodeLength(out); err != nil {
			return err
		} else {
			if len(*out) < int(val) {
				return fmt.Errorf(`input slice is short`)
			}
			t.SetString(string((*out)[:val]))
			*out = (*out)[val:]
		}
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			BinUnmarshal(out, t.Field(i).Addr().Interface())
		}
	case reflect.Slice:
		if val, err := DecodeLength(out); err != nil {
			return err
		} else {
			if len(*out) < int(val) {
				return fmt.Errorf(`input slice is short`)
			}
			t.SetBytes((*out)[:val])
			*out = (*out)[val:]
		}
	default:
		return fmt.Errorf(`unsupported type of BinUnmarshal %v`, t.Kind())
	}
	return nil
}

func FieldToBytes(v interface{}, num int) []byte {
	t := reflect.ValueOf(v)
	ret := make([]byte, 0, 2048)
	if t.Kind() == reflect.Struct && num < t.NumField() {
		field := t.Field(num)
		switch field.Kind() {
		case reflect.Uint8, reflect.Uint32, reflect.Uint64:
			ret = append(ret, []byte(fmt.Sprintf("%d", field.Uint()))...)
		case reflect.Int8, reflect.Int32, reflect.Int64:
			ret = append(ret, []byte(fmt.Sprintf("%d", field.Int()))...)
		case reflect.String:
			ret = append(ret, []byte(field.String())...)
		case reflect.Slice:
			ret = append(ret, field.Bytes()...)
			//		case reflect.Ptr:
			//		case reflect.Struct:
			//		default:
		}
	}
	return ret
}

func HexToInt64(input string) (ret int64) {
	hex, _ := hex.DecodeString(input)
	if length := len(hex); length <= 8 {
		ret = int64(binary.BigEndian.Uint64(append(make([]byte, 8-length), hex...)))
	}
	return
}
