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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testItem struct {
	Input        []int64
	Filter       string
	ParamsLength int
}

func TestGetNotificationCountFilter(t *testing.T) {
	testTable := []testItem{
		testItem{
			Input:        []int64{3, 5},
			Filter:       ` WHERE closed = false AND recipient_id IN (?) `,
			ParamsLength: 1,
		},
		testItem{
			Input:        nil,
			Filter:       ` WHERE closed = false `,
			ParamsLength: 0,
		},
	}

	for i, item := range testTable {
		filter, params := getNotificationCountFilter(item.Input, 1)
		assert.Equal(t, item.Filter, filter, "on %d step wrong filter %s", i, filter)
		assert.Equal(t, item.ParamsLength, len(params), "on %d step wrong params length %d", i, len(params))
	}

}
