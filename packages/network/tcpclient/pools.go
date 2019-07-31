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
	"sync"
)

// return nearest power of 2 that bigest than v
func powerOfTwo(v int) int64 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return int64(v)
}

var BytesPool *bytePool

func init() {
	BytesPool = &bytePool{
		pools: make(map[int64]*sync.Pool),
	}
}

type bytePool struct {
	pools map[int64]*sync.Pool
}

func (p *bytePool) Get(size int64) []byte {
	power := powerOfTwo(int(size))
	if pool, ok := p.pools[power]; ok {
		return pool.Get().([]byte)
	}

	pool := &sync.Pool{
		New: func() interface{} { return make([]byte, power) },
	}

	p.pools[power] = pool
	return pool.Get().([]byte)
}

func (p *bytePool) Put(buf []byte) {
	if len(buf) == 0 || buf == nil {
		return
	}

	if pool, ok := p.pools[int64(len(buf))]; ok {
		pool.Put(buf)
	}
}
