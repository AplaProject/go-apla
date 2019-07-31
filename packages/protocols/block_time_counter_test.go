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

package protocols

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockTimeCounter(t *testing.T) {
	btc := BlockTimeCounter{
		start:       time.Unix(0, 0),
		duration:    5 * time.Second,
		numberNodes: 3,
	}

	at := time.Unix(13, 0)

	queue, err := btc.queue(at)
	assert.NoError(t, err)
	assert.Equal(t, 2, queue)

	np, err := btc.nodePosition(at)
	assert.NoError(t, err)
	assert.Equal(t, 2, np)

	nextTime, err := btc.nextTime(at, 2)
	assert.NoError(t, err)
	assert.Equal(t, time.Unix(25, 0).Add(1*time.Millisecond), nextTime)

	start, end, err := btc.RangeByTime(at)
	assert.NoError(t, err)
	assert.Equal(t, time.Unix(10, 0).Add(1*time.Millisecond), start)
	assert.Equal(t, time.Unix(15, 0), end)
	fmt.Println("ranges:", start.Unix(), end.Unix())
}

func TestRangeByTime(t *testing.T) {
	btc := BlockTimeCounter{
		start:       time.Unix(1532977623, 0),
		duration:    4 * time.Second,
		numberNodes: 1,
	}

	st, end, err := btc.RangeByTime(time.Unix(1533062723, 0))
	require.NoError(t, err)
	fmt.Println(st.Unix(), end.Unix())

	st, end, err = btc.RangeByTime(time.Unix(1533062724, 0))
	require.NoError(t, err)
	fmt.Println(st.Unix(), end.Unix())

	// 1532977623
	st, end, err = btc.RangeByTime(time.Unix(1532977624, 0))
	require.NoError(t, err)
	fmt.Println(st.Unix(), end.Unix())

	// 1533062719 1533062723
	// 1533062723 1533062727
	// 1532977623 1532977627
}
