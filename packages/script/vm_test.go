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

package script

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcMem(t *testing.T) {
	cases := []struct {
		v   interface{}
		mem int64
	}{
		{true, 1},
		{int8(1), 1}, {int16(1), 2}, {int32(1), 4},
		{int64(1), 8}, {int(1), 8},
		{float32(1), 4}, {float64(1), 8},
		{"test", 4},
		{[]byte("test"), 16},
		{[]string{"test", "test"}, 20},
		{map[string]string{"test": "test"}, 12},
	}

	for _, v := range cases {
		assert.Equal(t, v.mem, calcMem(v.v))
	}
}
