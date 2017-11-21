package model

// LogTransaction is model
type LogTransaction struct {
	Hash []byte `gorm:"primary_key;not null"`
	Time int64  `gorm:"not null"`
}

// GetByHash returns LogTransactions existence by hash
func (lt *LogTransaction) GetByHash(hash []byte) (bool, error) {
	return isFound(DBConn.Where("hash = ?", hash).First(lt))
}

// Create is creating record of model
func (lt *LogTransaction) Create(transaction *DbTransaction) error {
	return GetDB(transaction).Create(lt).Error
}

// DeleteLogTransactionsByHash is deleting record by hash
func DeleteLogTransactionsByHash(transaction *DbTransaction, hash []byte) (int64, error) {
	query := GetDB(transaction).Exec("DELETE FROM log_transactions WHERE hash = ?", hash)
	return query.RowsAffected, query.Error
}

// GetLogTransactionsCount count records by transaction hash
func GetLogTransactionsCount(hash []byte) (int64, error) {
	var rowsCount int64
	if err := DBConn.Table("log_transactions").Where("hash = ?", hash).Count(&rowsCount).Error; err != nil {
		return -1, err
	}
	return rowsCount, nil
}
