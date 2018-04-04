package syspar

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFullNode(t *testing.T) {
	cases := []struct{ value, err string }{
		{`[[]]`, `Invalid format of the full_node parameter`},
		{`[["","","","", ""]]`, `Invalid values of the full_node parameter`},
		{`[["127.0.0.1", "https://127.0.0.1", "100", "c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7c7", "0"]]`, `Invalid values of the full_node parameter`},
		{`[["127.0.0.1", "https://127.0.0.1", "0", "c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7c7d0", "0"]]`, `Invalid values of the full_node parameter`},
		{`[["127.0.0.1", "https://", "100", "c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7c7d0", "0"]]`, `Invalid host: https://`},
		{`[["127.0.0.1", "127.0.0.1", "100", "c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7c7d0", "0"]]`, `parse 127.0.0.1: invalid URI for request`},
		{`[["", "https://127.0.0.1", "100", "c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7c7d0", "0"]]`, `Invalid values of the full_node parameter`},
		{`[["127.0.0.1", "https://127.0.0.1", "100", "c1a9e7b2fb8cea2a272e183c3e27e2d59a3ebe613f51873a46885c9201160bd263ef43b583b631edd1284ab42483712fd2ccc40864fe9368115ceeee47a7c7d0", "12345678"]]`, ``},
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
