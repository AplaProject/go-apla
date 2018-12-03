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

package tcpserver

import (
	"bytes"
	"fmt"

	"testing"

	"reflect"

	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/stretchr/testify/require"
)

func TestReadRequest(t *testing.T) {
	type testStruct struct {
		Id   uint32
		Data []byte
	}

	request := &bytes.Buffer{}
	request.Write(converter.DecToBin(10, 4))
	request.Write(converter.DecToBin(len("test"), 4))
	request.Write([]byte("test"))

	test := &testStruct{}
	err := ReadRequest(test, request)
	if err != nil {
		t.Errorf("read request return err: %s", err)
	}
	if test.Id != 10 {
		t.Errorf("bad id value")
	}
	if string(test.Data) != "test" {
		t.Errorf("bad data value: %+v", string(test.Data))
	}
}

func TestReadRequestTag(t *testing.T) {
	type testStruct2 struct {
		Id   uint32
		Data []byte `size:"4"`
	}

	request := &bytes.Buffer{}
	request.Write(converter.DecToBin(10, 4))
	request.Write([]byte("test"))

	test := &testStruct2{}
	err := ReadRequest(test, request)
	if err != nil {
		t.Errorf("read request return err: %s", err)
	}
	if test.Id != 10 {
		t.Errorf("bad id value")
	}
	if string(test.Data) != "test" {
		t.Errorf("bad data value: %+v", string(test.Data))
	}
}

func TestSendRequest(t *testing.T) {
	type testStruct2 struct {
		Id   uint32
		Id2  int64
		Test []byte
		Text []byte `size:"4"`
	}

	test := testStruct2{
		Id:   15,
		Id2:  0x1BCDEF0010203040,
		Test: []byte("test"),
		Text: []byte("text"),
	}

	bin := bytes.Buffer{}
	err := SendRequest(&test, &bin)
	if err != nil {
		t.Fatalf("send request failed: %s", err)
	}

	test2 := testStruct2{}
	err = ReadRequest(&test2, &bin)
	if err != nil {
		t.Fatalf("read request failed: %s", err)
	}

	if !reflect.DeepEqual(test, test2) {
		t.Errorf("different values: %+v and %+v", test, test2)
	}
}

func TestRequestType(t *testing.T) {
	source := RequestType{
		Type: uint16(RequestTypeNotFullNode),
	}

	buf := bytes.Buffer{}
	require.NoError(t, source.Write(&buf))

	target := RequestType{}
	require.NoError(t, target.Read(&buf))
	require.Equal(t, source.Type, target.Type)
}

func TestGetBodiesRequest(t *testing.T) {
	source := GetBodiesRequest{
		BlockID:      33,
		ReverseOrder: true,
	}

	buf := bytes.Buffer{}
	require.NoError(t, source.Write(&buf))

	target := GetBodiesRequest{}
	require.NoError(t, target.Read(&buf))
	fmt.Printf("%+v %+v\n", source, target)
	require.Equal(t, source, target)
}
