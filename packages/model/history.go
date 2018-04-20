package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/shopspring/decimal"
)

const historyTableSuffix = "_history"

var errLowBalance = errors.New("not enough APL on the balance")

// History represent record of history table
type History struct {
	tableName   string
	ID          int64
	SenderID    int64
	RecipientID int64
	Amount      decimal.Decimal
	Comment     string
	BlockID     int64
	TxHash      []byte `gorm:"column:txhash"`
	CreatedAt   time.Time
}

// SetTablePrefix is setting table prefix
func (h *History) SetTablePrefix(prefix int64) *History {
	h.tableName = HistoryTableName(prefix)
	return h
}

// TableName returns table name
func (h *History) TableName() string {
	return h.tableName
}

// APLTransfer from to amount
type APLTransfer struct {
	SenderID    int64
	RecipientID int64
	Amount      decimal.Decimal
}

//APLSenderTxCount struct to scan query result
type APLSenderTxCount struct {
	SenderID int64
	TxCount  int64
}

// GetExcessCommonTokenMovementPerDay returns sum of amounts 24 hours
func GetExcessCommonTokenMovementPerDay(tx *DbTransaction) (amount decimal.Decimal, err error) {
	db := GetDB(tx)
	type result struct {
		Amount decimal.Decimal
	}

	var res result
	err = db.Table("1_history").Select("SUM(amount) as amount").
		Where("created_at > NOW() - interval '24 hours' AND amount > 0").Scan(&res).Error

	return res.Amount, err
}

// GetExcessFromToTokenMovementPerDay returns from to pairs where sum of amount greather than fromToPerDayLimit per 24 hours
func GetExcessFromToTokenMovementPerDay(tx *DbTransaction) (excess []APLTransfer, err error) {
	db := GetDB(tx)
	err = db.Table("1_history").
		Select("sender_id, recipient_id, SUM(amount) amount").
		Where("created_at > NOW() - interval '24 hours' AND amount > 0").
		Group("sender_id, recipient_id").
		Having("SUM(amount) > ?", consts.FromToPerDayLimit).
		Scan(&excess).Error

	return excess, err
}

// GetExcessTokenMovementQtyPerBlock returns from to pairs where APL transactions count greather than tokenMovementQtyPerBlockLimit per 24 hours
func GetExcessTokenMovementQtyPerBlock(tx *DbTransaction, blockID int64) (excess []APLSenderTxCount, err error) {
	db := GetDB(tx)
	err = db.Table("1_history").
		Select("sender_id, count(*) tx_count").
		Where("block_id = ? AND amount > ?", blockID, 0).
		Group("sender_id").
		Having("count(*) > ?", consts.TokenMovementQtyPerBlockLimit).
		Scan(&excess).Error

	return excess, err
}

// HistoryTableName returns name of history table
func HistoryTableName(prefix int64) string {
	return fmt.Sprintf("%d%s", prefix, historyTableSuffix)
}
