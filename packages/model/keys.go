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

package model

import (
	"strconv"

	"github.com/shopspring/decimal"
)

// Key is model
type Key struct {
	ID          int64           `json:"id"`
	EcosystemID int64           `json:"ecosystem_id"`
	PublicKey   []byte          `json:"public_key"`
	Amount      decimal.Decimal `json:"amount"`
	Maxpay      string          `json:"maxpay"`
	Multi       bool            `json:"multi"`
	Deleted     bool            `json:"deleted"`
	Blocked     bool            `json:"blocked"`
}

func (k *Key) PrimaryKey() string {
	return "1_keys:" + strconv.FormatInt(k.EcosystemID, 10) +
		":" + strconv.FormatInt(k.ID, 10)
}

// Get is retrieving model from database
func (k *Key) Get(ecosystemID, id int64) (bool, error) {
	k.ID = id
	k.EcosystemID = ecosystemID
	return MetaStorage.Begin(false).FindModel(k)
}
