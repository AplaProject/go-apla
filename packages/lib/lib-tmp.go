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
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"reflect"
)

/*const (
	UpdPublicKey = `fd7f6ccf79ec35a7cf18640e83f0bbc62a5ae9ea7e9260e3a93072dd088d3c7acf5bcb95a7b44fcfceff8de4b16591d146bb3dc6e79f93f900e59a847d2684c3`
)*/

// Update contains version info parameters
type Update struct {
	Version string
	Hash    string
	Sign    string
	URL     string
}

// CalculateMd5 calculates MD5 hash of the file
func CalculateMd5(filePath string) ([]byte, error) {
	var result []byte
	file, err := os.Open(filePath)
	if err != nil {
		return result, err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return result, err
	}

	return hash.Sum(result), nil
}

// FieldToBytes returns the value of n-th field of v as []byte
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
		case reflect.Float64:
			ret = append(ret, []byte(fmt.Sprintf("%f", field.Float()))...)
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

//HexToInt64 converts hex int64 to int64
func HexToInt64(input string) (ret int64) {
	hex, _ := hex.DecodeString(input)
	if length := len(hex); length <= 8 {
		ret = int64(binary.BigEndian.Uint64(append(make([]byte, 8-length), hex...)))
	}
	return
}
