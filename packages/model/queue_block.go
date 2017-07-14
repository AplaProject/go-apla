package model

type QueueBlocks struct {
	Hash       []byte `gorm:"primary_key;not null"`
	BlockID    int64  `gorm:"not null"`
	FullNodeID int64  `gorm:"not null"`
}

func (q *QueueBlocks) GetQueueBlock() error {
	return DBConn.First(&q).Error
}

func (q *QueueBlocks) Delete() error {
	return DBConn.Delete(q).Error
}
