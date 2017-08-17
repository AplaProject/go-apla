package model

import (
	"strconv"
)

type WalletedTransaction struct {
	ID                     int64
	SenderWalletID         int64
	RecepientWalletID      int64
	RecepientWalletAddress string
	Amount                 string
	Comission              string
	Time                   int32
	Comment                string
	BlockID                int64
	RbID                   int64
	Sw                     int64
	Rw                     int64
}

func (wt *WalletedTransaction) Get(senderWalletID, recipientWalletID int64, recipientWalletAddress string, limit int, offset int) ([]WalletedTransaction, error) {
	result := new([]WalletedTransaction)
	err := DBConn.Table("dlt_transactions as d").Select("d.*, w.wallet_id as sw, wr.wallet_id as rw").
		Joins("left join dlt_wallets as w on w.wallet_id=d.sender_wallet_id").
		Joins("left join dlt_wallets as wr on wr.wallet_id=d.recipient_wallet_id").
		Where("sender_wallet_id=?", senderWalletID).
		Or("recipient_wallet_id=?", recipientWalletID).
		Or("recipient_wallet_address=?", recipientWalletAddress).
		Limit(limit).
		Offset(offset).
		Order("d.id desc").Scan(result).Error
	return *result, err
}

func (wt *WalletedTransaction) ToMap() map[string]string {
	result := make(map[string]string, 0)
	result["id"] = strconv.FormatInt(wt.ID, 10)
	result["sender_wallet_id"] = strconv.FormatInt(wt.SenderWalletID, 10)
	result["recepient_wallet_id"] = strconv.FormatInt(wt.RecepientWalletID, 10)
	result["recepient_wallet_address"] = wt.RecepientWalletAddress
	result["amount"] = wt.Amount
	result["comission"] = wt.Comission
	result["time"] = strconv.FormatInt(int64(wt.Time), 10)
	result["comment"] = wt.Comment
	result["block_id"] = strconv.FormatInt(wt.BlockID, 10)
	result["rb_id"] = strconv.FormatInt(wt.RbID, 10)
	result["sw"] = strconv.FormatInt(wt.Sw, 10)
	result["rw"] = strconv.FormatInt(wt.Rw, 10)
	return result
}
