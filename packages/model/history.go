package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

const (
	fromToPerDayLimit             = 10000
	tokenMovementQtyPerBlockLimit = 100
)

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
	if prefix == 0 {
		prefix = 1
	}
	h.tableName = fmt.Sprintf("%d_history", prefix)
	return h
}

// TableName returns table name
func (h *History) TableName() string {
	if h.tableName == "" {
		h.tableName = "1_history"
	}

	return h.tableName
}

// APLTransfer from to amount
type APLTransfer struct {
	SenderID    int64
	RecipientID int64
	Amount      float64
}

// GetExcessCommonTokenMovementPerDay returns sum of amounts 24 hours
func GetExcessCommonTokenMovementPerDay(tx *DbTransaction) (amount float64, err error) {
	db := GetDB(tx)
	type result struct {
		Amount float64
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
		Select("sender_id, recipient_id, SUM(amount) sum_amount").
		Where("created_at > NOW() - interval '24 hours' AND amount > 0").
		Group("sender_id, recipient_id").
		Having("SUM(amount) > ?", fromToPerDayLimit).
		Scan(&excess).Error

	return excess, err
}

// GetExcessTokenMovementQtyPerBlock returns from to pairs where APL transactions count greather than tokenMovementQtyPerBlockLimit per 24 hours
func GetExcessTokenMovementQtyPerBlock(tx *DbTransaction, blockID int64) (excess []APLTransfer, err error) {
	db := GetDB(tx)
	err = db.Table("1_history").
		Select("sender_id, count(*)").
		Where("block_id = ? AND amount > ?", blockID, 0).
		Group("sender_id").
		Having("count(*) > ?", tokenMovementQtyPerBlockLimit).
		Scan(&excess).Error

	return excess, err
}
