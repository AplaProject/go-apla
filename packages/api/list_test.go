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

package api

import (
	"fmt"
	"testing"

	"github.com/AplaProject/go-apla/packages/converter"
)

func TestList(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	var ret listResult
	err := sendGet(`list/contracts`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}
	if converter.StrToInt64(ret.Count) < 7 {
		t.Error(fmt.Errorf(`The number of records %s < 7`, ret.Count))
		return
	}
	err = sendGet(`list/qwert`, nil, &ret)
	if err.Error() != `400 {"error":"E_TABLENOTFOUND","msg":"Table qwert has not been found"}` {
		t.Error(err)
		return
	}
	var retTable tableResult
	for _, item := range []string{`app_params`, `parameters`} {
		err = sendGet(`table/`+item, nil, &retTable)
		if err != nil {
			t.Error(err)
			return
		}
		if retTable.Name != item {
			t.Errorf(`wrong table name %s != %s`, retTable.Name, item)
			return
		}
	}
	var sec listResult
	err = sendGet(`sections`, nil, &sec)
	if err != nil {
		t.Error(err)
		return
	}
	if converter.StrToInt(sec.Count) == 0 {
		t.Errorf(`section error`)
		return
	}
}
