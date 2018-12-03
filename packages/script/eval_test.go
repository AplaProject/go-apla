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

package script

import (
	"fmt"
	"testing"
)

type TestComp struct {
	Input  string
	Output string
}

func Multi(a, b int64) (int64, error) {
	return a + b*2, nil
}

func TestEvalIf(t *testing.T) {
	test := []TestComp{
		{`Multi(45, $citizenId")`, `there is not pair`},
		{"34 + `45` < 0", `runtime panic error`},
		{"Multi( (34+35)*2, Multi( $citizenId, 56))== 1 || Multi( (34+35)*2, Multi( $citizenId, 56))== 0", `false`},
		{"5 + 9 > 10", `true`},
		{"34 == 45", `false`},
		{"1345", `true`},
		{"13/13-1", `false`},
		{"7665 > ($citizenId-48000)", "false"},
		{"56788 + 1 >= $citizenId", "true"},
		{"76 < $citizenId", "true"},
		{"56789 <= $citizenId", "true"},
		{"56 == 56", "true"},
		{"37 != 37", "false"},
		{"!!(1-1)", "false"},
		{"!!$citizenId || $wallet_id", "true"},
		{"!789", "false"},
		{"$citizenId == 56780 + 9", `true`},
		{"qwerty(45)", `unknown identifier qwerty`},
		{"Multi(2, 5) > 36", "false"},
		{"789 63 == 63", "true"},
		{"+421", "stack is empty"},
		{"1256778+223445==1480223", "true"},
		{"(67-34789)*3 == -104166", "true"},
		{"(5+78)*(1563-527) == 85988", "true"},
		{"124 * (143-527", "there is not pair"},
		{"341 * 234/0", "divided by zero"},
		{"0 == ((15+82)*2 + 5)/2 - 99", "true"},
		{"Multi( (34+35)*2, Multi( $citizenId, 56))== 1 || Multi( (34+35)*2, Multi( $citizenId, 56))== 0", `false`},
		{"2+ Multi( (34+35)*2, Multi( $citizenId, 56)) /2 == 56972", `true`},
		{"$citizenId && 0", "false"},
		{"0|| ($citizenId + $wallet_id == 950240)", "true"},
	}
	vars := map[string]interface{}{
		`citizenId`: 56789,
		`wallet_id`: 893451,
	}
	vm := NewVM()
	vm.Extend(&ExtendData{map[string]interface{}{"Multi": Multi}, nil, nil})
	for _, item := range test {
		out, err := vm.EvalIf(item.Input, 0, &vars)
		if err != nil {
			if err.Error() != item.Output {
				t.Error(`error of ifeval ` + item.Input + ` ` + err.Error())
			}
		} else {
			if fmt.Sprint(out) != item.Output {
				t.Error(`error of ifeval ` + item.Input + ` Output:` + fmt.Sprint(out))
			}
		}
	}
}
