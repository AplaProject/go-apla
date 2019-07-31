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

package syspar

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFullNode(t *testing.T) {
	cases := []struct {
		value,
		err string
		formattingErr bool
	}{
		{value: `[{"tcp_address":"127.0.0.1", "api_address":"https://127.0.0.1", "key_id":"100", "public_key":"c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7c7d0", "unban_time": 111111}]`, err: ``},
		{value: `[{"tcp_address":"", "api_address":"https://127.0.0.1", "key_id":"100", "public_key":"c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7c7d0", "unban_time": 111111}]`, err: `Invalid values of the full_node parameter`},
		{value: `[{"tcp_address":"127.0.0.1", "api_address":"127.0.0.1", "key_id":"100", "public_key":"c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7c7d0", "unban_time": 111111}]`, err: `parse 127.0.0.1: invalid URI for request`},
		{value: `[{"tcp_address":"127.0.0.1", "api_address":"https://", "key_id":"100", "public_key":"c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7c7d0", "unban_time": 111111}]`, err: `Invalid host: https://`},
		{value: `[{"tcp_address":"127.0.0.1", "api_address":"https://127.0.0.1", "key_id":"0", "public_key":"c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7c7d0", "unban_time": 111111}]`, err: `Invalid values of the full_node parameter`},
		{value: `[{"tcp_address":"127.0.0.1", "api_address":"https://127.0.0.1", "key_id":"100", "public_key":"c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7c7d00000000000", "unban_time": 111111}]`, err: `Invalid values of the full_node parameter`},
		{value: `[{}}]`, err: `invalid character '}' after array element`, formattingErr: true},
	}
	for _, v := range cases {
		// Testing Unmarshalling string -> struct
		var fs []*FullNode
		err := json.Unmarshal([]byte(v.value), &fs)
		if len(v.err) == 0 {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, v.err)
		}

		// Testing Marshalling struct -> string
		blah, err := json.Marshal(fs)
		require.NoError(t, err)

		// Testing Unmarshaling string (from struct) -> struct
		var unfs []FullNode
		err = json.Unmarshal(blah, &unfs)
		if !v.formattingErr && len(v.err) != 0 {
			assert.EqualError(t, err, v.err)
		}
	}
}
