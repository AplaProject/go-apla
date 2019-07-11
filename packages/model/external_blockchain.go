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
	"strings"

	"github.com/AplaProject/go-apla/packages/converter"
)

// ExternalBlockchain represents a txinfo table
type ExternalBlockchain struct {
	Id               int64  `gorm:"primary_key;not null"`
	Value            string `gorm:"not null"`
	ExternalContract string `gorm:"not null"`
	ResultContract   string `gorm:"not null"`
	Url              string `gorm:"not null"`
	Uid              string `gorm:"not null"`
	TxTime           int64  `gorm:"not null"`
	Sent             int64  `gorm:"not null"`
	Hash             []byte `gorm:"not null"`
	Attempts         int64  `gorm:"not null"`
}

// GetExternalList returns the list of network tx
func GetExternalList() (list []ExternalBlockchain, err error) {
	err = DBConn.Table("external_blockchain").
		Order("id").Scan(&list).Error
	return
}

// DelExternalList deletes sent tx
func DelExternalList(list []int64) error {
	slist := make([]string, len(list))
	for i, v := range list {
		slist[i] = converter.Int64ToStr(v)
	}
	return DBConn.Exec("delete from external_blockchain where id in (" +
		strings.Join(slist, `,`) + ")").Error
}

func HashExternalTx(id int64, hash []byte) error {
	return DBConn.Exec("update external_blockchain set hash=?, sent = 1 where id = ?", hash, id).Error
}

func IncExternalAttempt(id int64) error {
	return DBConn.Exec("update external_blockchain set attempts=attempts+1 where id = ?", id).Error
}
