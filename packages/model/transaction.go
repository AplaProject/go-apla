// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package model

import (
	"github.com/AplaProject/go-apla/packages/consts"
)

// This constants contains values of transactions priority
const (
	TransactionRateOnBlock transactionRate = iota + 1
	TransactionRateStopNetwork
)

type transactionRate int8

// Transaction is model
type Transaction struct {
	Hash     []byte          `gorm:"private_key;not null"`
	Data     []byte          `gorm:"not null"`
	Used     int8            `gorm:"not null"`
	HighRate transactionRate `gorm:"not null"`
	Type     int8            `gorm:"not null"`
	KeyID    int64           `gorm:"not null"`
	Sent     int8            `gorm:"not null"`
	Verified int8            `gorm:"not null;default:1"`
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
func GetAllUnusedTransactions(limit int) ([]*Transaction, error) {
	var transactions []*Transaction

	query := DBConn.Where("used = ?", "0").Order("high_rate DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&transactions).Error; err != nil {
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
func DeleteTransactionIfUnused(transaction *DbTransaction, transactionHash []byte) (int64, error) {
	query := GetDB(transaction).Exec("DELETE FROM transactions WHERE hash = ? and used = 0 and verified = 0", transactionHash)
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
	if t.HighRate == 0 {
		t.HighRate = getTxRateByTxType(t.Type)
	}

	return DBConn.Create(t).Error
}

func getTxRateByTxType(txType int8) transactionRate {
	switch txType {
	case consts.TxTypeStopNetwork:
		return TransactionRateStopNetwork
	default:
		return 0
	}
}

func GetManyTransactions(dbtx *DbTransaction, hashes [][]byte) ([]Transaction, error) {
	txes := []Transaction{}
	query := GetDB(dbtx).Where("hash in (?)", hashes).Find(&txes)
	if err := query.Error; err != nil {
		return nil, err
	}

	return txes, nil
}
