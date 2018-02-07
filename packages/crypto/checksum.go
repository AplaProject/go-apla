// MIT License
//
// Copyright (c) 2016 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package crypto

import "hash/crc64"

type checksumProvider int

const (
	_CRC64 checksumProvider = iota
)

var (
	table64 *crc64.Table
)

func init() {
	table64 = crc64.MakeTable(crc64.ECMA)
}

// CalcChecksum is calculates checksum
func CalcChecksum(input []byte) (uint64, error) {
	switch checksumProv {
	case _CRC64:
		return calcCRC64(input), nil
	default:
		return 0, ErrUnknownProvider
	}
}

// CRC64 returns crc64 sum
func calcCRC64(input []byte) uint64 {
	return crc64.Checksum(input, table64)
}

// CheckSum calculates the 0-9 check sum of []byte
func checkSum(val []byte) int {
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
