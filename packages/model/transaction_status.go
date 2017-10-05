package model

import (
	"encoding/hex"
)

type TransactionStatus struct {
	Hash     []byte `gorm:"primary_key;not null"`
	Time     int64  `gorm:"not null;"`
	Type     int64  `gorm:"not null"`
	WalletID int64  `gorm:"not null"`
	BlockID  int64  `gorm:"not null"`
	Error    string `gorm:"not null;size 255"`
}

func (ts *TransactionStatus) TableName() string {
	return "transactions_status"
}

func (ts *TransactionStatus) Create() error {
	return DBConn.Create(ts).Error
}

func toHex(transactionHash []byte) []byte {
	for _, b := range transactionHash {
		if !(b >= '0' && b <= '9') && !(b >= 'a' && b <= 'f') {
			return []byte(hex.EncodeToString(transactionHash))
		}
	}
	return transactionHash
}

func (ts *TransactionStatus) Get(transactionHash []byte) (bool, error) {
	query := DBConn.Where("hash = ?", toHex(transactionHash)).First(ts)
	return query.RecordNotFound(), query.Error
}

func (ts *TransactionStatus) UpdateBlockID(transaction *DbTransaction, newBlockID int64, transactionHash []byte) error {
	return GetDB(transaction).Model(&TransactionStatus{}).Where("hash = ?", transactionHash).Update("block_id", newBlockID).Error
}

func (ts *TransactionStatus) UpdateBlockMsg(newBlockID int64, msg string, transactionHash []byte) error {
	return DBConn.Model(&TransactionStatus{}).Where("hash = ?", toHex(transactionHash)).Updates(
		map[string]interface{}{"block_id": newBlockID, "error": msg}).Error
}

func (ts *TransactionStatus) SetError(errorText string, transactionHash []byte) error {
	return DBConn.Model(&TransactionStatus{}).Where("hash = ?", toHex(transactionHash)).Update("error", errorText).Error
}
