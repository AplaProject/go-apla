package model

// QueueTx is model
type QueueTx struct {
	Hash     []byte `gorm:"primary_key;not null"`
	Data     []byte `gorm:"not null"`
	FromGate int    `gorm:"not null"`
}

// TableName returns name of table
func (qt *QueueTx) TableName() string {
	return "queue_tx"
}

// DeleteTx is deleting tx
func (qt *QueueTx) DeleteTx() error {
	return DBConn.Delete(qt).Error
}

// Save is saving model
func (qt *QueueTx) Save(transaction *DbTransaction) error {
	return GetDB(transaction).Save(qt).Error
}

// Create is creating record of model
func (qt *QueueTx) Create() error {
	return DBConn.Create(qt).Error
}

// GetByHash is retrieving model from database by hash
func (qt *QueueTx) GetByHash(hash []byte) (bool, error) {
	return isFound(DBConn.Where("hash = ?", hash).First(qt))
}

// DeleteQueueTxByHash is deleting queue tx by hash
func DeleteQueueTxByHash(transaction *DbTransaction, hash []byte) (int64, error) {
	query := GetDB(transaction).Exec("DELETE FROM queue_tx WHERE hash = ?", hash)
	return query.RowsAffected, query.Error
}

// GetQueuedTransactionsCount counting queued transactions
func GetQueuedTransactionsCount(hash []byte) (int64, error) {
	var rowsCount int64
	err := DBConn.Table("queue_tx").Where("hash = ?", hash).Count(&rowsCount).Error
	return rowsCount, err
}

// GetAllUnverifiedAndUnusedTransactions is returns all unverified and unused transaction
func GetAllUnverifiedAndUnusedTransactions() ([]*QueueTx, error) {
	query := `SELECT *
		  FROM (
	              SELECT data,
	                     hash
	              FROM queue_tx
		      UNION
		      SELECT data,
			     hash
		      FROM transactions
		      WHERE verified = 0 AND used = 0
			)  AS x`
	rows, err := DBConn.Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var data, hash []byte
	result := []*QueueTx{}
	for rows.Next() {
		if err := rows.Scan(&data, &hash); err != nil {
			return nil, err
		}
		result = append(result, &QueueTx{Data: data, Hash: hash})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
