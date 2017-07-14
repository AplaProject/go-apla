package model

type Transactions struct {
	Hash      []byte `gorm:private_key;not null`
	Data      []byte `gorm:not null`
	Used      int8   `gorm:not null`
	HighRate  int8   `gorm:not null`
	Type      int8   `gorm:not null`
	WalletID  int64  `gorm:not null`
	CitizeniD int64  `gorm:not null`
	Counter   int8   `gorm:not null`
	Sent      int8   `gorm:not null`
}

func GetAllUnusedAndVerifiedTransactions() (*[]Transactions, error) {
	transactions := new([]Transactions)
	if err := DBConn.Where("used = ? AND verified = ?", "0", "1").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func GetAllUnsendedAndUnselfTransactions() (*[]Transactions, error) {
	transactions := new([]Transactions)
	if err := DBConn.Where("sent = ? AND for_self_use = ?", "0", "0").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func DeleteLoopedTransactions() (int64, error) {
	query := DBConn.Exec("DELETE FROM transactions WHERE verified = 0 AND used = 0 AND counter > 10")
	return query.RowsAffected, query.Error
}

func (t *Transactions) Read(hash []byte) error {
	return DBConn.Where("hash = ?", hash).First(t).Error
}

func GetAllUnsendedTransactions() (*[]Transactions, error) {
	transactions := new([]Transactions)
	if err := DBConn.Where("sent = ?", "0").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (t *Transactions) Save() error {
	return DBConn.Save(t).Error
}

func MarkTransactionsUnverified() (int64, error) {
	query := DBConn.Exec("UPDATE transactions SET verified = 0 WHERE verified = 1 AND used = 0")
	return query.RowsAffected, query.Error
}

func MarkTransactionSended(transactionHash []byte) (int64, error) {
	query := DBConn.Exec("UPDATE transactions SET sent = 1 WHERE hex(hash) = ?", transactionHash)
	return query.RowsAffected, query.Error
}
