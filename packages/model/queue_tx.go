package model

type QueueTx struct {
	Hash     []byte `gorm:"primary_key;not null"`
	Data     []byte `gorm:"not null"`
	FromGate int    `gorm:"not null"`
}

func (qt *QueueTx) TableName() string {
	return "queue_tx"
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

func DeleteQueueTxByHash(hash []byte) (int64, error) {
	query := DBConn.Exec("DELETE FROM queue_tx WHERE hex(hash) = ?", hash)
	return query.RowsAffected, query.Error
}

func DeleteQueuedTransaction(hash []byte) error {
	return DBConn.Exec("DELETE FROM queue_tx WHERE hex(hash) = ?", hash).Error
}

func GetQueuedTransactionsCount(hash []byte) (int64, error) {
	var rowsCount int64
	err := DBConn.Where("hash = ?", hash).Count(&rowsCount).Error
	return rowsCount, err
}
