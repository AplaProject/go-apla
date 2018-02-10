// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package model

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

// GetBlockchain is retrieving chain of blocks from database
func GetBlockchain(startBlockID int64, endblockID int64) ([]Block, error) {
	var err error
	blockchain := new([]Block)
	if endblockID > 0 {
		err = DBConn.Model(&Block{}).Order("id asc").Where("id > ? AND id <= ?", startBlockID, endblockID).Find(&blockchain).Error
	} else {
		err = DBConn.Model(&Block{}).Order("id asc").Where("id > ?", startBlockID).Find(&blockchain).Error
	}
	if err != nil {
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
func (b *Block) GetBlocksFrom(startFromID int64, ordering string) ([]Block, error) {
	var err error
	blockchain := new([]Block)
	err = DBConn.Order("id "+ordering).Where("id > ?", startFromID).Find(&blockchain).Error
	return *blockchain, err
}

// DeleteByID is deleting block by ID
func (b *Block) DeleteByID(transaction *DbTransaction, id int64) error {
	return GetDB(transaction).Where("id = ?", id).Delete(Block{}).Error
}
