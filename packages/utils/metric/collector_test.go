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

package metric

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func MockValue(v int64) *Value {
	return &Value{Time: 1, Metric: "test_metric", Key: "ecosystem_1", Value: v}
}

func MockCollectorFunc(v int64, err error) CollectorFunc {
	return func() ([]*Value, error) {
		if err != nil {
			return nil, err
		}

		return []*Value{MockValue(v)}, nil
	}
}

func TestValue(t *testing.T) {
	value := MockValue(100)
	result := map[string]interface{}{"time": int64(1), "metric": "test_metric", "key": "ecosystem_1", "value": int64(100)}
	assert.Equal(t, result, value.ToMap())
}

func TestCollector(t *testing.T) {
	c := NewCollector(
		MockCollectorFunc(100, nil),
		MockCollectorFunc(0, errors.New("Test")),
		MockCollectorFunc(200, nil),
	)

	result := []interface{}{
		map[string]interface{}{"time": int64(1), "metric": "test_metric", "key": "ecosystem_1", "value": int64(100)},
		map[string]interface{}{"time": int64(1), "metric": "test_metric", "key": "ecosystem_1", "value": int64(200)},
	}
	assert.Equal(t, result, c.Values())
}
