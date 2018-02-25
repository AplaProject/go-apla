package model

import "time"

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

func (b *Block) GetReverseBlockchain(endBlockID int64, limit int32) ([]Block, error) {
	var err error
	blockchain := new([]Block)
	err = DBConn.Model(&Block{}).Order("id DESC").Where("id <= ?", endBlockID).Limit(limit).Find(&blockchain).Error
	return *blockchain, err
}

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
