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
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
)

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

// UintToBytes converts 32-byte value into [4]byte (BigEndian)
func UintToBytes(val uint32) []byte {
	tmp := make([]byte, 4)
	binary.BigEndian.PutUint32(tmp, val)
	return tmp
}

// EncodeLength encodes int64 number to []byte. If it is less than 128 then it returns []byte{length}.
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

// DecodeLength decodes []byte to int64 and shifts buf. Bytes must be encoded with EncodeLength function.
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

// EncodeLenByte appends the length of the slice (EncodeLength) + the slice.
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

// Time32 gets the current time in UNIX format.
func Time32() uint32 {
	return uint32(time.Now().Unix())
}

// BinMarshal converts v parameter to []byte slice.
func BinMarshal(out *[]byte, v interface{}) (*[]byte, error) {
	var err error

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
			*out = append(*out, 128+4-i)
			*out = append(*out, tmp[i:]...)
		}
	case reflect.Float64:
		bin := Float2Bytes(t.Float())
		*out = append(*out, bin...)
	case reflect.Int64:
		EncodeLenInt64(out, t.Int())
	case reflect.Uint64:
		tmp := make([]byte, 8)
		binary.BigEndian.PutUint64(tmp, t.Uint())
		*out = append(*out, tmp...)
	case reflect.String:
		*out = append(append(*out, EncodeLength(int64(t.Len()))...), []byte(t.String())...)
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if out, err = BinMarshal(out, t.Field(i).Interface()); err != nil {
				return out, err
			}
		}
	case reflect.Slice:
		*out = append(append(*out, EncodeLength(int64(t.Len()))...), t.Bytes()...)
	case reflect.Ptr:
		if out, err = BinMarshal(out, t.Elem().Interface()); err != nil {
			return out, err
		}
	default:
		return out, fmt.Errorf(`unsupported type of BinMarshal`)
	}
	return out, nil
}

// BinUnmarshal converts []byte slice which has been made with BinMarshal to v
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
			if len(*out) <= int(size) || size > 4 {
				return fmt.Errorf(`wrong input data`)
			}
			for ; i < size; i++ {
				tmp[4-size+i] = (*out)[i+1]
			}
			t.SetInt(int64(binary.BigEndian.Uint32(tmp)))
			*out = (*out)[size+1:]
		}
	case reflect.Float64:
		t.SetFloat(Bytes2Float((*out)[:8]))
		*out = (*out)[8:]
	case reflect.Int64:
		val, err := DecodeLenInt64(out)
		if err != nil {
			return err
		}
		t.SetInt(val)
	case reflect.Uint64:
		t.SetUint(binary.BigEndian.Uint64((*out)[:8]))
		*out = (*out)[8:]
	case reflect.String:
		val, err := DecodeLength(out)
		if err != nil {
			return err
		}
		if len(*out) < int(val) {
			return fmt.Errorf(`input slice is short`)
		}
		t.SetString(string((*out)[:val]))
		*out = (*out)[val:]
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if err := BinUnmarshal(out, t.Field(i).Addr().Interface()); err != nil {
				return err
			}
		}
	case reflect.Slice:
		val, err := DecodeLength(out)
		if err != nil {
			return err
		}
		if len(*out) < int(val) {
			return fmt.Errorf(`input slice is short`)
		}
		t.SetBytes((*out)[:val])
		*out = (*out)[val:]
	default:
		return fmt.Errorf(`unsupported type of BinUnmarshal %v`, t.Kind())
	}
	return nil
}

// EscapeName deletes unaccessable characters for input name(s)
func EscapeName(name string) string {
	out := make([]byte, 1, len(name)+2)
	out[0] = '"'
	available := `() ,`
	for _, ch := range []byte(name) {
		if (ch >= '0' && ch <= '9') || ch == '_' || (ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') || strings.IndexByte(available, ch) >= 0 {
			out = append(out, ch)
		}
	}
	if strings.IndexAny(string(out), available) >= 0 {
		return string(out[1:])
	}
	return string(append(out, '"'))
}

// Escape deletes unaccessable characters
func Escape(data string) string {
	out := make([]byte, 0, len(data)+2)
	available := `_ ,=!-'()"?*$<>: `
	for _, ch := range []byte(data) {
		if (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') || strings.IndexByte(available, ch) >= 0 {
			out = append(out, ch)
		}
	}
	return string(out)
}

// EscapeForJSON replaces quote to slash and quote
func EscapeForJSON(data string) string {
	return strings.Replace(data, `"`, `\"`, -1)
}

// NumString insert spaces between each three digits. 7123456 => 7 123 456
func NumString(in string) string {
	if strings.IndexByte(in, '.') >= 0 {
		lr := strings.Split(in, `.`)
		return NumString(lr[0]) + `.` + lr[1]
	}
	buf := []byte(in)
	out := make([]byte, len(in)+4)
	for len(buf) > 3 {
		out = append(append([]byte(` `), buf[len(buf)-3:]...), out...)
		buf = buf[:len(buf)-3]
	}
	return string(append(buf, out...))
}

// Bytes2Float converts []byte to float64
func Bytes2Float(bytes []byte) float64 {
	return math.Float64frombits(binary.LittleEndian.Uint64(bytes))
}

// Float2Bytes converts float64 to []byte
func Float2Bytes(float float64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, math.Float64bits(float))
	return bytes
}

// StripTags replaces < and > to &lt; and &gt;
func StripTags(value string) string {
	return strings.Replace(strings.Replace(value, `<`, `&lt;`, -1), `>`, `&gt;`, -1)
}

// EGSMoney converts qEGS to EGS. For example, 123455000000000000000 => 123.455
func EGSMoney(money string) string {
	digit := consts.EGS_DIGIT
	if len(money) < digit+1 {
		money = strings.Repeat(`0`, digit+1-len(money)) + money
	}
	money = money[:len(money)-digit] + `.` + money[len(money)-digit:]
	return strings.TrimRight(strings.TrimRight(money, `0`), `.`)
}
