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
	"time"
)

// Block is model
type Block struct {
	ID            int64  `gorm:"primary_key;not_null"`
	Hash          []byte `gorm:"not null"`
	RollbacksHash []byte `gorm:"not null"`
	Data          []byte `gorm:"not null"`
	EcosystemID   int64  `gorm:"not null"`
	KeyID         int64  `gorm:"not null"`
	NodePosition  int64  `gorm:"not null"`
	Time          int64  `gorm:"not null"`
	Tx            int32  `gorm:"not null"`
}

// TableName returns name of table
func (Block) TableName() string {
	return "block_chain"
}

// Create is creating record of model
func (b *Block) Create(transaction *DbTransaction) error {
	return GetDB(transaction).Create(b).Error
}

// Get is retrieving model from database
func (b *Block) Get(blockID int64) (bool, error) {
	return isFound(DBConn.Where("id = ?", blockID).First(b))
}

// GetMaxBlock returns last block existence
func (b *Block) GetMaxBlock() (bool, error) {
	return isFound(DBConn.Last(b))
}

// GetMaxForeignBlock returns last block generated not by key_id
func (b *Block) GetMaxForeignBlock(keyId int64) (bool, error) {
	return isFound(DBConn.Order("id DESC").Where("key_id != ?", keyId).First(b))
}

// GetBlockchain is retrieving chain of blocks from database
func GetBlockchain(startBlockID int64, endblockID int64, order ordering) ([]Block, error) {
	var err error
	blockchain := new([]Block)

	orderStr := "id " + string(order)
	query := DBConn.Model(&Block{}).Order(orderStr)
	if endblockID > 0 {
		query = query.Where("id > ? AND id <= ?", startBlockID, endblockID).Find(&blockchain)
	} else {
		query = query.Where("id > ?", startBlockID).Find(&blockchain)
	}

	if query.Error != nil {
		return nil, err
	}
	return *blockchain, nil
}

// GetBlocks is retrieving limited chain of blocks from database
func (b *Block) GetBlocks(startFromID int64, limit int32) ([]Block, error) {
	var err error
	blockchain := new([]Block)
	if startFromID > 0 {
		err = DBConn.Order("id desc").Limit(limit).Where("id > ?", startFromID).Find(&blockchain).Error
	} else {
		err = DBConn.Order("id desc").Limit(limit).Find(&blockchain).Error
	}
	return *blockchain, err
}

// GetBlocksFrom is retrieving ordered chain of blocks from database
func (b *Block) GetBlocksFrom(startFromID int64, ordering string, limit int32) ([]Block, error) {
	var err error
	blockchain := new([]Block)
	if limit == 0 {
		err = DBConn.Order("id "+ordering).Where("id > ?", startFromID).Find(&blockchain).Error
	} else {
		err = DBConn.Order("id "+ordering).Where("id > ?", startFromID).Limit(limit).Find(&blockchain).Error
	}
	return *blockchain, err
}

// GetReverseBlockchain returns records of blocks in reverse ordering
func (b *Block) GetReverseBlockchain(endBlockID int64, limit int32) ([]Block, error) {
	var err error
	blockchain := new([]Block)
	err = DBConn.Model(&Block{}).Order("id DESC").Where("id <= ?", endBlockID).Limit(limit).Find(&blockchain).Error
	return *blockchain, err
}

// GetNodeBlocksAtTime returns records of blocks for time interval and position of node
func (b *Block) GetNodeBlocksAtTime(from, to time.Time, node int64) ([]Block, error) {
	var err error
	blockchain := new([]Block)
	err = DBConn.Model(&Block{}).Where("node_position = ? AND time BETWEEN ? AND ?", node, from.Unix(), to.Unix()).Find(&blockchain).Error
	return *blockchain, err
}

// DeleteById is deleting block by ID
func (b *Block) DeleteById(transaction *DbTransaction, id int64) error {
	return GetDB(transaction).Where("id = ?", id).Delete(Block{}).Error
}
