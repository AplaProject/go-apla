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

package test

import (
	"encoding/hex"
	"math/rand"
	"time"
)

type WantString struct {
	Input string
	Want  string
}

// Generates a random []bytes.
func RandBytes(length int) ([]byte, int64) {
	seed := time.Now().UTC().UnixNano()
	rng := rand.New(rand.NewSource(seed))
	ret := make([]byte, length)
	for ; length > 0; length-- {
		ret[length-1] = byte(rng.Intn(256))
	}
	return ret, seed
}

func HexToBytes(input string) []byte {
	ret, _ := hex.DecodeString(input)
	return ret
}
