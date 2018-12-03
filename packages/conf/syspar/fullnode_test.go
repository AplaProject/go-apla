// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.
//
// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.
//
// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

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
