package model

type LogTransactions struct {
	Hash []byte `gorm:"primary_key;not null"`
	Time int64  `gorm:"not null"`
}

func (lt *LogTransactions) IsExists() (bool, error) {
	query := DBConn.First(lt)
	return !query.RecordNotFound(), query.Error
}

func (lt *LogTransactions) Delete() error {
	return DBConn.Delete(lt).Error
}

func (lt *LogTransactions) Get() error {
	return DBConn.First(lt).Error
}

func (lt *LogTransactions) GetByHash(hash []byte) error {
	return DBConn.Where("hex(hash) = ?").First(lt).Error
}

func (lt *LogTransactions) Create() error {
	return DBConn.Create(lt).Error
}

func DeleteLogTransactionsByHash(hash []byte) (int64, error) {
	query := DBConn.Exec("DELETE FROM log_transactions WHERE hex(hash) = ?", hash)
	return query.RowsAffected, query.Error
}
