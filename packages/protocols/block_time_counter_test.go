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
