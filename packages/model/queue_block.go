package model

// QueueBlock is model
type QueueBlock struct {
	Hash       []byte `gorm:"primary_key;not null"`
	BlockID    int64  `gorm:"not null"`
	FullNodeID int64  `gorm:"not null"`
}

// Get is retrieving model from database
func (qb *QueueBlock) Get() (bool, error) {
	return isFound(DBConn.First(qb))
}

// GetQueueBlockByHash is retrieving blocks queue by hash
func (qb *QueueBlock) GetQueueBlockByHash(hash []byte) (bool, error) {
	return isFound(DBConn.Where("hash = ?", hash).First(qb))
}

// Delete is deleting queue
func (qb *QueueBlock) Delete() error {
	return DBConn.Delete(qb).Error
}

// DeleteQueueBlockByHash is deleting queue by hash
func (qb *QueueBlock) DeleteQueueBlockByHash() error {
	query := DBConn.Exec("DELETE FROM queue_blocks WHERE hash = ?", qb.Hash)
	return query.Error
}

// DeleteOldBlock is deleting old blocks
func (qb *QueueBlock) DeleteOldBlocks() error {
	query := DBConn.Exec("DELETE FROM queue_blocks WHERE block_id <= ?", qb.BlockID)
	return query.Error
}

// Create is creating record of model
func (qb *QueueBlock) Create() error {
	return DBConn.Create(qb).Error
}
