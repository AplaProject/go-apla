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
