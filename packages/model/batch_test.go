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

package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestBatchModel struct {
	ID   int64
	Name string
}

func (m TestBatchModel) TableName() string {
	return "test_batch"
}

func (m TestBatchModel) FieldValue(fieldName string) (interface{}, error) {
	switch fieldName {
	case "id":
		return m.ID, nil
	case "name":
		return m.Name, nil
	default:
		return nil, fmt.Errorf("Unknown field %s of TestBatchModel", fieldName)
	}
}

func TestPrepareQuery(t *testing.T) {
	slice := []BatchModel{
		TestBatchModel{ID: 1, Name: "first"},
		TestBatchModel{ID: 2, Name: "second"},
	}

	query, args, err := prepareQuery(slice, []string{"id", "name"})
	require.NoError(t, err)

	checkQuery := `INSERT INTO "test_batch" (id,name) VALUES (?,?),(?,?)`
	checkArgs := []interface{}{int64(1), "first", int64(2), "second"}

	require.Equal(t, checkQuery, query)
	require.Equal(t, checkArgs, args)
}
