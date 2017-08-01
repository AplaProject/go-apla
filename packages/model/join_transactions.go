package model

import "github.com/shopspring/decimal"

type WalletedTransaction struct {
	ID                     int64
	SenderWalletID         int64
	RecepientWalletID      int64
	RecepientWalletAddress string
	Amount                 decimal.Decimal
	Comission              decimal.Decimal
	Time                   int32
	Comment                string
	BlockID                int64
	RbID                   int64
	Sw                     int64
	Rw                     int64
}

func (wt *WalletedTransaction) Get(senderWalletID, recipientWalletID int64, recipientWalletAddress string, limit int, offset int) ([]WalletedTransaction, error) {
	var result []WalletedTransaction
	err := DBConn.Table("dlt_transactions as d").Select("d.*, w.wallet_id as sw, wr.wallet_id as rw").
		Joins("left join dlt_wallets as w on w.wallet_id=d.sender_wallet_id").
		Joins("left join dlt_wallets as wr on wr.wallet_id=d.recipient_wallet_id").
		Where("sender_wallet_id=?", senderWalletID).
		Or("recipient_wallet_id=?", recipientWalletID).
		Or("recipient_wallet_address=?", recipientWalletAddress).
		Limit(limit).
		Offset(offset).
		Order("d.id desc").Scan(wt).Error
	return result, err
}

func (wt *WalletedTransaction) ToMap() map[string]string {
	result := make(map[string]string, 0)
	result["id"] = string(wt.ID)
	result["sender_wallet_id"] = string(wt.SenderWalletID)
	result["recepient_wallet_id"] = string(wt.RecepientWalletID)
	result["recepient_wallet_address"] = wt.RecepientWalletAddress
	result["amount"] = wt.Amount.String()
	result["comission"] = wt.Comission.String()
	result["time"] = string(wt.Time)
	result["comment"] = wt.Comment
	result["block_id"] = string(wt.BlockID)
	result["rb_id"] = string(wt.RbID)
	result["sw"] = string(wt.Sw)
	result["rw"] = string(wt.Rw)
	return result
}
