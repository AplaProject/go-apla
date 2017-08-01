package model

type Transaction struct {
	Hash       []byte `gorm:private_key;not null`
	Data       []byte `gorm:not null`
	Used       int8   `gorm:not null`
	HighRate   int8   `gorm:not null`
	Type       int8   `gorm:not null`
	ForSelfUse int8   `gorm:not null`
	WalletID   int64  `gorm:not null`
	CitizenID  int64  `gorm:not null`
	ThirdVar   int32  `gorm:not null`
	Counter    int8   `gorm:not null`
	Sent       int8   `gorm:not null`
}

func GetAllUnusedTransactions() (*[]Transaction, error) {
	transactions := new([]Transaction)
	if err := DBConn.Where("used = ?", "0").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func GetAllUnsendedTransactions() (*[]Transaction, error) {
	transactions := new([]Transaction)
	if err := DBConn.Where("sent = ?", "0").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func DeleteLoopedTransactions() (int64, error) {
	query := DBConn.Exec("DELETE FROM transactions WHERE used = 0 AND counter > 10")
	return query.RowsAffected, query.Error
}

func DeleteTransactionByHash(hash []byte) (int64, error) {
	query := DBConn.Exec("DELETE FROM transactions WHERE hex(hash) = ?", hash)
	return query.RowsAffected, query.Error
}

func DeleteUsedTransactions() (int64, error) {
	query := DBConn.Exec("DELETE FROM transactions WHERE used = 1")
	return query.RowsAffected, query.Error
}

func MarkTransactionSended(transactionHash []byte) (int64, error) {
	query := DBConn.Exec("UPDATE transactions SET sent = 1 WHERE hex(hash) = ?", transactionHash)
	return query.RowsAffected, query.Error
}

func MarkTransactionUsed(transactionHash []byte) (int64, error) {
	query := DBConn.Exec("UPDATE transactions SET used = 1 WHERE hex(hash) = ?", transactionHash)
	return query.RowsAffected, query.Error
}

func MarkTransactionUnusedAndUnverified(transactionHash []byte) (int64, error) {
	query := DBConn.Exec("UPDATE transactions SET used = 0, verified = 0 WHERE hex(hash) = ?", transactionHash)
	return query.RowsAffected, query.Error
}

func MarkVerifiedAndNotUsedTransactionsUnverified() (int64, error) {
	query := DBConn.Exec("UPDATE transactions SET verified = 0 WHERE verified = 1 AND used = 0")
	return query.RowsAffected, query.Error
}

func MarkTransactionUnused(transactionHash []byte) (int64, error) {
	query := DBConn.Exec("UPDATE transactions SET used = 0 WHERE hex(hash) = ?", transactionHash)
	return query.RowsAffected, query.Error
}

func (t *Transaction) Read(hash []byte) error {
	return DBConn.Where("hash = ?", hash).First(t).Error
}

func (t *Transaction) Save() error {
	return DBConn.Save(t).Error
}

func GetLastTransactions(limit int) ([]Transaction, error) {
	transactions := new([]Transaction)
	if err := DBConn.Limit(limit).Find(transactions).Error; err != nil {
		return nil, err
	}
	return *transactions, nil
}

func (t *Transaction) IsExists() (bool, error) {
	query := DBConn.First(t)
	return !query.RecordNotFound(), query.Error
}

func (t *Transaction) DeleteBad() error {
	return DBConn.Where("used = ? and counter > ?", 0, 10).Delete(t).Error
}

func (t *Transaction) Get(transactionHash []byte) error {
	return DBConn.Where("hash = ?", transactionHash).First(t).Error
}

func (t *Transactions) GetVerified(transactionHash []byte) error {
	return DBConn.Where("hex(hash) = ? AND verified = 1", transactionHash).First(t).Error
}

func DeleteTransaction(transactionHash []byte) (int64, error) {
	query := DBConn.Exec("DELETE FROM transactions WHERE hash = ?", transactionHash)
	return query.RowsAffected, query.Error
}

func DeleteTransactionIfUnused(transactionHash []byte) (int64, error) {
	query := DBConn.Exec("DELETE FROM transactions WHERE hash = ? and used = 0", transactionHash)
	return query.RowsAffected, query.Error
}

func (t *Transaction) Create() error {
	return DBConn.Create(t).Error
}
