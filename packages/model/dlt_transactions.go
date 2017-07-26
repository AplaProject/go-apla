package model

import "github.com/shopspring/decimal"

type DltTransactions struct {
	ID                     int64           `gorm:"primary_key;not null"`
	SenderWalletID         int64           `gorm:"not null"`
	RecepientWalletID      int64           `gorm:"not null"`
	RecepientWalletAddress string          `gorm:"not null;size:32"`
	Amount                 decimal.Decimal `gorm:"not null"`
	Comission              decimal.Decimal `gorm:"not null"`
	Time                   int64           `gorm:"not null"`
	Comment                string          `gorm:"not null"`
	BlockID                int64           `gorm:"not null"`
	RbID                   int64           `gorm:"not null"`
}

func (dt *DltTransactions) Create() error {
	return DBConn.Create(dt).Error
}

func (dt *DltTransactions) GetTransaction(senderWalletID, recipientWalletID int64, recipientWalletAddress string) error {
	return DBConn.Where("sender_wallet_id = ? OR recipient_wallet_id = ? OR recipient_wallet_address = ?",
		senderWalletID, recipientWalletID, recipientWalletAddress).First(dt).Error
}

func (dt *DltTransactions) GetIncomingTransactions(recipientWalletID int64) error {
	return DBConn.Where("recipient_wallet_id=?", recipientWalletID).First(dt).Error
}

/*
func (db *DCDB) GetAllTxBySenderOrRecepient(senderWalletID, recipientWalletID int64, recipientWalletAddress string, limit string) ([]map[string]string, error) {
	return db.GetAll(`SELECT d.*, w.wallet_id as sw, wr.wallet_id as rw FROM dlt_transactions as d
		        left join dlt_wallets as w on w.wallet_id=d.sender_wallet_id
		        left join dlt_wallets as wr on wr.wallet_id=d.recipient_wallet_id
				where sender_wallet_id=? OR
		        recipient_wallet_id=?  OR
		        recipient_wallet_address=? order by d.id desc  `+limit, -1, senderWalletID, senderWalletID, recipientWalletAddress)
}
*/
