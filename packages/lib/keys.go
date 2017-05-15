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
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash/crc64"
	"math/big"
	"strconv"
	"strings"
)

var (
	table64 *crc64.Table
)

func init() {
	table64 = crc64.MakeTable(crc64.ECMA)
}

// AddressToString converts int64 address to EGAAS address as XXXX-...-XXXX.
func AddressToString(address int64) (ret string) {
	num := strconv.FormatUint(uint64(address), 10)
	val := []byte(strings.Repeat("0", 20-len(num)) + num)

	for i := 0; i < 4; i++ {
		ret += string(val[i*4:(i+1)*4]) + `-`
	}
	ret += string(val[16:])
	return
}

// StringToAddress converts string EGAAS address to int64 address. The input address can be a positive or negative
// number, or EGAAS address in XXXX-...-XXXX format. Returns 0 when error occurs.
func StringToAddress(address string) (result int64) {
	var (
		err error
		ret uint64
	)
	if len(address) == 0 {
		return 0
	}
	if address[0] == '-' {
		var id int64
		id, err = strconv.ParseInt(address, 10, 64)
		if err != nil {
			return 0
		}
		address = strconv.FormatUint(uint64(id), 10)
	}
	if len(address) < 20 {
		address = strings.Repeat(`0`, 20-len(address)) + address
	}

	val := []byte(strings.Replace(address, `-`, ``, -1))
	if len(val) != 20 {
		return
	}
	if ret, err = strconv.ParseUint(string(val), 10, 64); err != nil {
		return 0
	}
	if CheckSum(val[:len(val)-1]) != int(val[len(val)-1]-'0') {
		return 0
	}
	result = int64(ret)
	return
}

// FillLeft fills the slice by zero at left if the size of the slice is less than 32.
func FillLeft(slice []byte) []byte {
	if len(slice) >= 32 {
		return slice
	}
	return append(make([]byte, 32-len(slice)), slice...)
}

// FillLeft64 fills the slice by zero at left if the size of the slice is less than 32.
func FillLeft64(slice []byte) []byte {
	if len(slice) >= 64 {
		return slice
	}
	return append(make([]byte, 64-len(slice)), slice...)
}

// GenBytesKeys generates a random pair of ECDSA private and public binary keys.
func GenBytesKeys() ([]byte, []byte, error) {
	private, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return private.D.Bytes(), append(FillLeft(private.PublicKey.X.Bytes()), FillLeft(private.PublicKey.Y.Bytes())...), nil
}

// GenHexKeys generates a random pair of ECDSA private and public hex keys.
func GenHexKeys() (string, string, error) {
	priv, pub, err := GenBytesKeys()
	if err != nil {
		return ``, ``, err
	}
	return hex.EncodeToString(priv), hex.EncodeToString(pub), nil
}

// PrivateToPublic returns the public key for the specified private key.
func PrivateToPublic(key []byte) []byte {
	pubkeyCurve := elliptic.P256()
	bi := new(big.Int).SetBytes(key)
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = pubkeyCurve
	priv.D = bi
	priv.PublicKey.X, priv.PublicKey.Y = pubkeyCurve.ScalarBaseMult(bi.Bytes())
	return append(FillLeft(priv.PublicKey.X.Bytes()), FillLeft(priv.PublicKey.Y.Bytes())...)
}

// PrivateToPublicHex returns the hex public key for the specified hex private key.
func PrivateToPublicHex(hexkey string) string {
	key, err := hex.DecodeString(hexkey)
	if err != nil {
		return ``
	}
	return hex.EncodeToString(PrivateToPublic(key))
}

// CheckSum calculates the 0-9 check sum of []byte
func CheckSum(val []byte) int {
	var one, two int
	for i, ch := range val {
		digit := int(ch - '0')
		if i&1 == 1 {
			one += digit
		} else {
			two += digit
		}
	}
	checksum := (two + 3*one) % 10
	if checksum > 0 {
		checksum = 10 - checksum
	}
	return checksum
}

// IsValidAddress checks if the specified address is EGAAS address.
func IsValidAddress(address string) bool {
	val := []byte(strings.Replace(address, `-`, ``, -1))
	if len(val) != 20 {
		return false
	}
	if _, err := strconv.ParseUint(string(val), 10, 64); err != nil {
		return false
	}
	return CheckSum(val[:len(val)-1]) == int(val[len(val)-1]-'0')
}

// CRC64 returns crc64 sum
func CRC64(input []byte) uint64 {
	return crc64.Checksum(input, table64)
}

// Address gets int64 EGGAS address from the public key
func Address(pubKey []byte) int64 {
	h256 := sha256.Sum256(pubKey)
	h512 := sha512.Sum512(h256[:])
	crc := CRC64(h512[:])
	// replace the last digit by checksum
	num := strconv.FormatUint(crc, 10)
	val := []byte(strings.Repeat("0", 20-len(num)) + num)
	return int64(crc - (crc % 10) + uint64(CheckSum(val[:len(val)-1])))
}

// KeyToAddress converts a public key to EGAAS address XXXX-...-XXXX.
func KeyToAddress(pubKey []byte) string {
	return AddressToString(Address(pubKey))
}
