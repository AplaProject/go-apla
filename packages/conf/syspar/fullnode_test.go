package syspar

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFullNode(t *testing.T) {
	cases := []struct{ value, err string }{
		{`[{"tcp_address":"127.0.0.1", "api_address":"https://127.0.0.1", "key_id":"100", "public_key":"c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7c7d0"}]`, ``},
		{`[{"tcp_address":"", "api_address":"https://127.0.0.1", "key_id":"100", "public_key":"c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7c7d0"}]`, `Invalid values of the full_node parameter`},
		{`[{"tcp_address":"127.0.0.1", "api_address":"127.0.0.1", "key_id":"100", "public_key":"c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7c7d0"}]`, `parse 127.0.0.1: invalid URI for request`},
		{`[{"tcp_address":"127.0.0.1", "api_address":"https://", "key_id":"100", "public_key":"c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7c7d0"}]`, `Invalid host: https://`},
		{`[{"tcp_address":"127.0.0.1", "api_address":"https://127.0.0.1", "key_id":"0", "public_key":"c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7c7d0"}]`, `Invalid values of the full_node parameter`},
		{`[{"tcp_address":"127.0.0.1", "api_address":"https://127.0.0.1", "key_id":"100", "public_key":"c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7c7d00000000000"}]`, `Invalid values of the full_node parameter`},
		{`[{}}]`, `invalid character '}' after array element`},
	}
	for _, v := range cases {
		var fs []*FullNode

		err := json.Unmarshal([]byte(v.value), &fs)
		if len(v.err) == 0 {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, v.err)
		}

		blah, err := json.Marshal(fs)
		assert.NoError(t, err)

		var unfs []FullNode
		assert.NoError(t, json.Unmarshal(blah, &unfs))
	}
}
