//MIT License
//
//Copyright (c) 2016-2018 GenesisKernel
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.

package model

// Transaction is model
type Transaction struct {
	Hash     []byte `gorm:"private_key;not null"`
	Data     []byte `gorm:"not null"`
	Used     int8   `gorm:"not null"`
	HighRate int8   `gorm:"not null"`
	Type     int8   `gorm:"not null"`
	KeyID    int64  `gorm:"not null"`
	Counter  int8   `gorm:"not null"`
	Sent     int8   `gorm:"not null"`
	Verified int8   `gorm:"not null;default:1"`
}

// GetAllTransactions is retrieving all transactions with limit
func GetAllTransactions(limit int) (*[]Transaction, error) {
	transactions := new([]Transaction)
	if err := DBConn.Limit(limit).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetAllUnusedTransactions is retrieving all unused transactions
func GetAllUnusedTransactions() ([]Transaction, error) {
	var transactions []Transaction
	if err := DBConn.Where("used = ?", "0").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetAllUnsentTransactions is retrieving all unset transactions
func GetAllUnsentTransactions() (*[]Transaction, error) {
	transactions := new([]Transaction)
	if err := DBConn.Where("sent = ?", "0").Find(transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetTransactionCountAll count all transactions
func GetTransactionCountAll() (int64, error) {
	var rowsCount int64
	if err := DBConn.Table("transactions").Count(&rowsCount).Error; err != nil {
		return -1, err
	}
	return rowsCount, nil
}

// GetTransactionsCount count all transactions by hash
func GetTransactionsCount(hash []byte) (int64, error) {
	var rowsCount int64
	if err := DBConn.Table("transactions").Where("hash = ?", hash).Count(&rowsCount).Error; err != nil {
		return -1, err
	}
	return rowsCount, nil
}

// DeleteLoopedTransactions deleting lopped transactions
func DeleteLoopedTransactions() (int64, error) {
	query := DBConn.Exec("DELETE FROM transactions WHERE used = 0 AND counter > 10")
	return query.RowsAffected, query.Error
}

// DeleteTransactionByHash deleting transaction by hash
func DeleteTransactionByHash(hash []byte) (int64, error) {
	query := DBConn.Exec("DELETE FROM transactions WHERE hash = ?", hash)
	return query.RowsAffected, query.Error
}

// DeleteUsedTransactions deleting used transaction
func DeleteUsedTransactions(transaction *DbTransaction) (int64, error) {
	query := GetDB(transaction).Exec("DELETE FROM transactions WHERE used = 1")
	return query.RowsAffected, query.Error
}

// DeleteTransactionIfUnused deleting unused transaction
func DeleteTransactionIfUnused(transactionHash []byte) (int64, error) {
	query := DBConn.Exec("DELETE FROM transactions WHERE hash = ? and used = 0 and verified = 0", transactionHash)
	return query.RowsAffected, query.Error
}

// MarkTransactionSent is marking transaction as sent
func MarkTransactionSent(transactionHash []byte) (int64, error) {
	query := DBConn.Exec("UPDATE transactions SET sent = 1 WHERE hash = ?", transactionHash)
	return query.RowsAffected, query.Error
}

// MarkTransactionUsed is marking transaction as used
func MarkTransactionUsed(transaction *DbTransaction, transactionHash []byte) (int64, error) {
	query := GetDB(transaction).Exec("UPDATE transactions SET used = 1 WHERE hash = ?", transactionHash)
	return query.RowsAffected, query.Error
}

// MarkTransactionUnusedAndUnverified is marking transaction unused and unverified
func MarkTransactionUnusedAndUnverified(transaction *DbTransaction, transactionHash []byte) (int64, error) {
	query := GetDB(transaction).Exec("UPDATE transactions SET used = 0, verified = 0 WHERE hash = ?", transactionHash)
	return query.RowsAffected, query.Error
}

// MarkVerifiedAndNotUsedTransactionsUnverified is marking verified and unused transaction as unverified
func MarkVerifiedAndNotUsedTransactionsUnverified() (int64, error) {
	query := DBConn.Exec("UPDATE transactions SET verified = 0 WHERE verified = 1 AND used = 0")
	return query.RowsAffected, query.Error
}

// Read is checking transaction existence by hash
func (t *Transaction) Read(hash []byte) (bool, error) {
	return isFound(DBConn.Where("hash = ?", hash).First(t))
}

// Get is retrieving model from database
func (t *Transaction) Get(transactionHash []byte) (bool, error) {
	return isFound(DBConn.Where("hash = ?", transactionHash).First(t))
}

// GetVerified is checking transaction verification by hash
func (t *Transaction) GetVerified(transactionHash []byte) (bool, error) {
	return isFound(DBConn.Where("hash = ? AND verified = 1", transactionHash).First(t))
}

// Create is creating record of model
func (t *Transaction) Create() error {
	return DBConn.Create(t).Error
}
