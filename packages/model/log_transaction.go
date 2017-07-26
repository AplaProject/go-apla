package model

type LogTransactions struct {
	Hash []byte `gorm:"primary_key;not null"`
	Time int32  `gorm:"not null"`
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

func (lt *LogTransactions) Create() error {
	return DBConn.Create(lt).Error
}

func GetLogTransactionsCount(hash []byte) (int64, error) {
	var rowsCount int64
	if err := DBConn.Exec("SELECT count(hash) FROM log_transactions WHERE hex(hash) = ?", hash).Scan(&rowsCount).Error; err != nil {
		return -1, err
	}
	return rowsCount, nil
}
