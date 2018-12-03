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
