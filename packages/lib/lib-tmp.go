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

/*
//HexToInt64 converts hex int64 to int64
func HexToInt64(input string) (ret int64) {
	hex, _ := hex.DecodeString(input)
	if length := len(hex); length <= 8 {
		ret = int64(binary.BigEndian.Uint64(append(make([]byte, 8-length), hex...)))
	}
	return
}
*/
