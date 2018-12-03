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

package querycost

import (
	"github.com/AplaProject/go-apla/packages/model"
)

type QueryCosterType int

const (
	ExplainQueryCosterType        QueryCosterType = iota
	ExplainAnalyzeQueryCosterType QueryCosterType = iota
	FormulaQueryCosterType        QueryCosterType = iota
)

type QueryCoster interface {
	QueryCost(*model.DbTransaction, string, ...interface{}) (int64, error)
}

type ExplainQueryCoster struct {
}

func (*ExplainQueryCoster) QueryCost(transaction *model.DbTransaction, query string, args ...interface{}) (int64, error) {
	return explainQueryCost(transaction, true, query, args...)
}

type ExplainAnalyzeQueryCoster struct {
}

func (*ExplainAnalyzeQueryCoster) QueryCost(transaction *model.DbTransaction, query string, args ...interface{}) (int64, error) {
	return explainQueryCost(transaction, true, query, args...)
}

func GetQueryCoster(tp QueryCosterType) QueryCoster {
	switch tp {
	case ExplainQueryCosterType:
		return &ExplainQueryCoster{}
	case ExplainAnalyzeQueryCosterType:
		return &ExplainAnalyzeQueryCoster{}
	case FormulaQueryCosterType:
		return &FormulaQueryCoster{&DBCountQueryRowCounter{}}
	}
	return nil
}
