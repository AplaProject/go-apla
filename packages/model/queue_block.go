package model

type QueueBlock struct {
	Hash       []byte `gorm:"primary_key;not null"`
	BlockID    int64  `gorm:"not null"`
	FullNodeID int64  `gorm:"not null"`
}

func (qb *QueueBlock) GetQueueBlock() error {
	return DBConn.First(&qb).Error
}

func (qb *QueueBlock) Delete() error {
	return DBConn.Delete(qb).Error
}

func (qb *QueueBlock) Create() error {
	return DBConn.Create(qb).Error
}

func InsertQueueBlock(hash []byte, fullNodeID int64, blockID int64) error {
	return DBConn.Exec(`INSERT INTO queue_blocks(hash, full_node_id, block_id) VALUES ([hex], ?, ?) ON CONFLICT DO NOTHING`,
		hash, fullNodeID, blockID).Error

}
