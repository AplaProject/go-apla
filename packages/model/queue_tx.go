package model

type QueueTx struct {
	Hash     []byte `gorm:"primary_key;not null"`
	Data     []byte `gorm:"not null"`
	FromGate int    `gorm:"not null"`
}

func DeleteQueueTx() error {
	return DBConn.Delete(&QueueTx{}).Error
}

func (qt *QueueTx) DeleteTx() error {
	return DBConn.Delete(qt).Error
}

func (qt *QueueTx) Save() error {
	return DBConn.Save(qt).Error
}
