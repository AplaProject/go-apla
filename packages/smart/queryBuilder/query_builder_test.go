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

package queryBuilder

import (
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
)

// query="SELECT ,,,id,amount,\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\"ecosystem\"
// FROM \"1_keys\" \nWHERE  AND id = -6752330173818123413 AND ecosystem = '1'\n"

// fields="[+amount]"
// values="[2912910000000000000]"

// whereF="[id]"
// whereV="[-6752330173818123413]"

type TestKeyTableChecker struct {
	Val bool
}

func (tc TestKeyTableChecker) IsKeyTable(tableName string) bool {
	return tc.Val
}
func TestSqlFields(t *testing.T) {
	qb := smartQueryBuilder{
		Entry:        log.WithFields(log.Fields{"mod": "test"}),
		table:        "1_keys",
		Fields:       []string{"+amount"},
		FieldValues:  []interface{}{2912910000000000000},
		WhereFields:  []string{"id"},
		WhereValues:  []string{"-6752330173818123413"},
		KeyTableChkr: TestKeyTableChecker{true},
	}

	fields, err := qb.GetSQLSelectFieldsExpr()
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(fields)
}
