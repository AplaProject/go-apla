package model

type LogTransaction struct {
	Hash []byte `gorm:"primary_key;not null"`
	Time int64  `gorm:"not null"`
}

func (lt *LogTransaction) IsExists() (bool, error) {
	query := DBConn.First(lt)
	return !query.RecordNotFound(), query.Error
}

func (lt *LogTransaction) Delete() error {
	return DBConn.Delete(lt).Error
}

func (lt *LogTransaction) Get() error {
	return DBConn.First(lt).Error
}

func (lt *LogTransaction) GetByHash(hash []byte) error {
	return DBConn.Where("hex(hash) = ?").First(lt).Error
}

func (lt *LogTransaction) Create() error {
	return DBConn.Create(lt).Error
}

func DeleteLogTransactionsByHash(hash []byte) (int64, error) {
	query := DBConn.Exec("DELETE FROM log_transactions WHERE hex(hash) = ?", hash)
	return query.RowsAffected, query.Error
}

func GetLogTransactionsCount(hash []byte) (int64, error) {
	var rowsCount int64
	if err := DBConn.Exec("SELECT count(hash) FROM log_transactions WHERE hex(hash) = ?", hash).Scan(&rowsCount).Error; err != nil {
		return -1, err
	}
	return rowsCount, nil
}

func LogTransactionsCreateTable() error {
	return DBConn.CreateTable(&LogTransaction{}).Error
}
