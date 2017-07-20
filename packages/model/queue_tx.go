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

func (qt *QueueTx) Create() error {
	return DBConn.Create(qt).Error
}

func (qt *QueueTx) GetByHash(hash []byte) error {
	return DBConn.Where("hex(hash) = ?").First(qt).Error
}

func GetQueuedTransactionsCount() (int64, error) {
	var rowsCount int64
	if err := DBConn.Exec("SELECT count(hash) FROM queue_tx WHERE hex(hash) = ?").Scan(&rowsCount).Error; err != nil {
		return -1, err
	}
	return rowsCount, nil
}
