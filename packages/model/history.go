// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package model

import (
	"errors"
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/shopspring/decimal"
)

var errLowBalance = errors.New("not enough APL on the balance")

// History represent record of history table
type History struct {
	ecosystem   int64
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
	h.ecosystem = prefix
	return h
}

// TableName returns table name
func (h *History) TableName() string {
	if h.ecosystem == 0 {
		h.ecosystem = 1
	}
	return `1_history`
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
		Where("to_timestamp(created_at) > NOW() - interval '24 hours' AND amount > 0").Scan(&res).Error

	return res.Amount, err
}

// GetExcessFromToTokenMovementPerDay returns from to pairs where sum of amount greather than fromToPerDayLimit per 24 hours
func GetExcessFromToTokenMovementPerDay(tx *DbTransaction) (excess []APLTransfer, err error) {
	db := GetDB(tx)
	err = db.Table("1_history").
		Select("sender_id, recipient_id, SUM(amount) amount").
		Where("to_timestamp(created_at) > NOW() - interval '24 hours' AND amount > 0").
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
