package model

import "github.com/jinzhu/gorm"

type Transaction struct {
	Hash      []byte `gorm:"private_key;not null"`
	Data      []byte `gorm:"not null"`
	Used      int8   `gorm:"not null"`
	HighRate  int8   `gorm:"not null"`
	Type      int8   `gorm:"not null"`
	WalletID  int64  `gorm:"not null"`
	CitizenID int64  `gorm:"not null"`
	Counter   int8   `gorm:"not null"`
	Sent      int8   `gorm:"not null"`
	Verified  int8   `gorm:"not null;default:1"`
}

func GetAllUnusedTransactions() (*[]Transaction, error) {
	transactions := new([]Transaction)
	if err := DBConn.Where("used = ?", "0").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func GetAllTransactions(limit int) (*[]Transaction, error) {
	transactions := new([]Transaction)
	if err := DBConn.Limit(limit).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func GetAllUnsentTransactions() (*[]Transaction, error) {
	transactions := new([]Transaction)
	if err := DBConn.Where("sent = ?", "0").Find(transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func GetLastTransactions(limit int) ([]Transaction, error) {
	transactions := new([]Transaction)
	if err := DBConn.Limit(limit).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return *transactions, nil
}

func GetTransactionsCount(hash []byte) (int64, error) {
	var rowsCount int64
	if err := DBConn.Table("transactions").Where("hash = ?", hash).Count(&rowsCount).Error; err != nil {
		return -1, err
	}
	return rowsCount, nil
}

func GetTransactionCountAll() (int64, error) {
	var rowsCount int64
	if err := DBConn.Table("transactions").Count(&rowsCount).Error; err != nil {
		return -1, err
	}
	return rowsCount, nil
}

func DeleteLoopedTransactions() (int64, error) {
	query := DBConn.Exec("DELETE FROM transactions WHERE used = 0 AND counter > 10")
	return query.RowsAffected, query.Error
}

func DeleteTransactionByHash(hash []byte) (int64, error) {
	query := DBConn.Exec("DELETE FROM transactions WHERE hash = ?", hash)
	return query.RowsAffected, query.Error
}

func DeleteUsedTransactions(transaction *DbTransaction) (int64, error) {
	query := GetDB(transaction).Exec("DELETE FROM transactions WHERE used = 1")
	return query.RowsAffected, query.Error
}

func DeleteTransactionIfUnused(transactionHash []byte) (int64, error) {
	query := DBConn.Exec("DELETE FROM transactions WHERE hash = ? and used = 0 and verified = 0", transactionHash)
	return query.RowsAffected, query.Error
}

func MarkTransactionSent(transactionHash []byte) (int64, error) {
	query := DBConn.Exec("UPDATE transactions SET sent = 1 WHERE hash = ?", transactionHash)
	return query.RowsAffected, query.Error
}

func MarkTransactionUsed(transaction *DbTransaction, transactionHash []byte) (int64, error) {
	query := GetDB(transaction).Exec("UPDATE transactions SET used = 1 WHERE hash = ?", transactionHash)
	return query.RowsAffected, query.Error
}

func MarkTransactionUnusedAndUnverified(transaction *DbTransaction, transactionHash []byte) (int64, error) {
	query := GetDB(transaction).Exec("UPDATE transactions SET used = 0, verified = 0 WHERE hash = ?", transactionHash)
	return query.RowsAffected, query.Error
}

func MarkVerifiedAndNotUsedTransactionsUnverified() (int64, error) {
	query := DBConn.Exec("UPDATE transactions SET verified = 0 WHERE verified = 1 AND used = 0")
	return query.RowsAffected, query.Error
}

func MarkTransactionUnused(transactionHash []byte) (int64, error) {
	query := DBConn.Exec("UPDATE transactions SET used = 0 WHERE hash = ?", transactionHash)
	return query.RowsAffected, query.Error
}

func (t *Transaction) Read(hash []byte) error {
	return handleError(DBConn.Where("hash = ?", hash).First(t).Error)
}

func (t *Transaction) Save() error {
	return DBConn.Save(t).Error
}

func (t *Transaction) Get(transactionHash []byte) error {
	return handleError(DBConn.Where("hash = ?", transactionHash).First(t).Error)
}

func (t *Transaction) GetVerified(transactionHash []byte) error {
	return handleError(DBConn.Where("hash = ? AND verified = 1", transactionHash).First(t).Error)
}

func (t *Transaction) IsExists() (bool, error) {
	query := DBConn.First(t)
	if query.Error == gorm.ErrRecordNotFound {
		return false, nil
	}
	return !query.RecordNotFound(), query.Error
}

func (t *Transaction) Create() error {
	return DBConn.Create(t).Error
}

func TransactionsCreateTable() error {
	return DBConn.CreateTable(&Transaction{}).Error
}
