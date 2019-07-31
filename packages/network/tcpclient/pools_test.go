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

package tcpclient

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBytesPoolGet(t *testing.T) {

	buf := BytesPool.Get(12832256)
	require.Equal(t, 16777216, len(buf))
}

func TestBytesPoolPut(t *testing.T) {
	short := []byte(strings.Repeat("A", 5))
	buf := BytesPool.Get(12832256)
	copy(buf[:5], short)
	BytesPool.Put(buf)

	newBuf := BytesPool.Get(12832256)
	require.Equal(t, 16777216, len(newBuf))

	require.Equal(t, newBuf[:5], short)
	fmt.Println(newBuf[:6])
}

func TestBytesPoolCicle(t *testing.T) {
	short := []byte(strings.Repeat("A", 5))
	buf := BytesPool.Get(int64(len(short)))
	copy(buf[:5], short)
	BytesPool.Put(buf)

	power := powerOfTwo(5)
	fmt.Println("power", power)

	newBuf := BytesPool.Get(5)
	require.Equal(t, power, int64(len(newBuf)))

	require.Equal(t, newBuf[:5], short)
	fmt.Println(newBuf)
}
