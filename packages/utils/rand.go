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

package utils

import (
	"math/rand"

	"github.com/AplaProject/go-apla/packages/crypto"
)

type Rand struct {
	src *rand.Rand
}

func (r *Rand) BytesSeed(b []byte) *rand.Rand {
	seed, _ := crypto.CalcChecksum(b)
	r.src.Seed(int64(seed))
	return r.src
}

func NewRand(seed int64) *Rand {
	return &Rand{
		src: rand.New(rand.NewSource(seed)),
	}
}
