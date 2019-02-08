// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

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

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmptyGetBodyResponse(t *testing.T) {
	buf := []byte{}
	w := bytes.NewBuffer(buf)
	empty := &GetBodyResponse{}
	require.NoError(t, empty.Write(w))

	r := bytes.NewReader(w.Bytes())
	emptyRes := &GetBodyResponse{}
	require.NoError(t, emptyRes.Read(r))
}

func TestWriteReadInts(t *testing.T) {
	buf := []byte{}
	b := bytes.NewBuffer(buf)
	st := uint16(2)
	require.NoError(t, binary.Write(b, binary.LittleEndian, st))

	var val uint16
	err := binary.Read(b, binary.LittleEndian, &val)
	require.NoError(t, err)
	require.Equal(t, val, st)
	fmt.Println(val)
}

func TestRequestType(t *testing.T) {
	rt := RequestType{1}
	buf := []byte{}
	b := bytes.NewBuffer(buf)

	result := RequestType{}
	require.NoError(t, rt.Write(b))
	require.NoError(t, result.Read(b))
	require.Equal(t, rt, result)
	fmt.Println(rt, result)

}

func TestGetBodyResponse(t *testing.T) {
	rt := GetBodyResponse{Data: make([]byte, 4, 4)}
	buf := []byte{}
	b := bytes.NewBuffer(buf)

	result := GetBodyResponse{}
	require.NoError(t, rt.Write(b))
	require.NoError(t, result.Read(b))
	require.Equal(t, rt, result)
	fmt.Println(rt, result)

}

func TestBodyResponse(t *testing.T) {
	rt := GetBodyResponse{Data: []byte(strings.Repeat("A", 32))}
	buf := []byte{}
	b := bytes.NewBuffer(buf)

	result := &GetBodyResponse{}
	require.NoError(t, rt.Write(b))
	require.NoError(t, result.Read(b))
	require.Equal(t, rt.Data, result.Data)
	fmt.Println(rt, result)

}
