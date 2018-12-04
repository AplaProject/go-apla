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
	"github.com/AplaProject/go-apla/packages/converter"
)

// InfoBlock is model
type InfoBlock struct {
	Hash           []byte `gorm:"not null"`
	EcosystemID    int64  `gorm:"not null default 0"`
	KeyID          int64  `gorm:"not null default 0"`
	NodePosition   string `gorm:"not null default 0"`
	BlockID        int64  `gorm:"not null"`
	Time           int64  `gorm:"not null"`
	CurrentVersion string `gorm:"not null"`
	Sent           int8   `gorm:"not null"`
}

// TableName returns name of table
func (ib *InfoBlock) TableName() string {
	return "info_block"
}

// Get is retrieving model from database
func (ib *InfoBlock) Get() (bool, error) {
	return isFound(DBConn.Last(ib))
}

// Update is update model
func (ib *InfoBlock) Update(transaction *DbTransaction) error {
	return GetDB(transaction).Model(&InfoBlock{}).Updates(ib).Error
}

// GetUnsent is retrieving model from database
func (ib *InfoBlock) GetUnsent() (bool, error) {
	return isFound(DBConn.Where("sent = ?", "0").First(&ib))
}

// Create is creating record of model
func (ib *InfoBlock) Create(transaction *DbTransaction) error {
	return GetDB(transaction).Create(ib).Error
}

// MarkSent update model sent field
func (ib *InfoBlock) MarkSent() error {
	return DBConn.Model(ib).Update("sent", 1).Error
}

// BlockGetUnsent returns InfoBlock
func BlockGetUnsent() (*InfoBlock, error) {
	ib := &InfoBlock{}
	found, err := ib.GetUnsent()
	if !found {
		return nil, err
	}
	return ib, err
}

// Marshall returns block as []byte
func (ib *InfoBlock) Marshall() []byte {
	if ib != nil {
		toBeSent := converter.DecToBin(ib.BlockID, 3)
		return append(toBeSent, ib.Hash...)
	}
	return []byte{}
}
