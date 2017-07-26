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

func DeleteQueuedTransaction(hash []byte) error {
	return DBConn.Exec("DELETE FROM queue_tx WHERE hex(hash) = ?", hash).Error
}

func GetQueuedTransactionsCount(hash []byte) (int64, error) {
	var rowsCount int64
	if err := DBConn.Exec("SELECT count(hash) FROM queue_tx WHERE hex(hash) = ?", hash).Scan(&rowsCount).Error; err != nil {
		return -1, err
	}
	return rowsCount, nil
}

func InsertIntoQueueTransaction(hash, data []byte, fromGate int) error {
	return DBConn.Exec("INSERT INTO queue_tx (hash, data, from_gate) VALUES ([hex], [hex], ?)", hash, data, fromGate).Error
}
