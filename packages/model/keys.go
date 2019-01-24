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

import "fmt"

// Key is model
type Key struct {
	ecosystem int64
	ID        int64  `gorm:"primary_key;not null"`
	PublicKey []byte `gorm:"column:pub;not null"`
	Amount    string `gorm:"not null"`
	Maxpay    string `gorm:"not null"`
	Deleted   int64  `gorm:"not null"`
	Blocked   int64  `gorm:"not null"`
	ReadOnly  int64  `gorm:"not null"`
}

// SetTablePrefix is setting table prefix
func (m *Key) SetTablePrefix(prefix int64) *Key {
	m.ecosystem = prefix
	return m
}

// TableName returns name of table
func (m Key) TableName() string {
	if m.ecosystem == 0 {
		m.ecosystem = 1
	}
	return `1_keys`
}

// Get is retrieving model from database
func (m *Key) Get(wallet int64) (bool, error) {
	return isFound(DBConn.Where("id = ? and ecosystem = ?", wallet, m.ecosystem).First(m))
}

// KeyTableName returns name of key table
func KeyTableName(prefix int64) string {
	return fmt.Sprintf("%d_keys", prefix)
}

// GetKeysCount returns common count of keys
func GetKeysCount() (int64, error) {
	var cnt int64
	row := DBConn.Raw(`SELECT count(*) key_count FROM "1_keys" WHERE ecosystem = 1`).Select("key_count").Row()
	err := row.Scan(&cnt)
	return cnt, err
}
