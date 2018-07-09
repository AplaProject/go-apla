package model

import (
	"fmt"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	log "github.com/sirupsen/logrus"
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
	Counter  int8            `gorm:"not null"`
	Sent     int8            `gorm:"not null"`
	Attempt  int8            `gorm:"not null"`
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

// IncrementTxAttemptCount increases attempt column
func IncrementTxAttemptCount(transaction *DbTransaction, transactionHash []byte) (int64, error) {
	defer func() {
		go logTrBigAttemptCount(transaction, transactionHash)
	}()

	query := GetDB(transaction).Exec("update transactions set attempt=attempt+1, used = case when attempt>10 then 1 else 0 end where hash = ?",
		transactionHash)
	return query.RowsAffected, query.Error
}

func getTxRateByTxType(txType int8) transactionRate {
	switch txType {
	case consts.TxTypeStopNetwork:
		return TransactionRateStopNetwork
	default:
		return 0
	}
}

func logTrBigAttemptCount(tbtx *DbTransaction, txHash []byte) {
	t := Transaction{}
	if err := tbtx.conn.Where("hash = ?", txHash).First(&t).Error; err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on getting tx by hash")
		return
	}

	if t.Attempt >= 10 {
		txString := fmt.Sprintf("tx_hash: %s, tx_data: %s, tx_attempt: %d", t.Hash, t.Data, t.Attempt)
		log.WithFields(log.Fields{"type": consts.BadTxError, "tx_info": txString}).Error("logging tx attempt count")
	}
}
