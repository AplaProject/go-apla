package model

import "github.com/shopspring/decimal"

type DltTransaction struct {
	ID                     int64            `gorm:"primary_key;not null"`
	SenderWalletID         int64            `gorm:"not null"`
	RecipientWalletID      int64            `gorm:"not null"`
	RecipientWalletAddress string           `gorm:"not null;size:32"`
	Amount                 *decimal.Decimal `gorm:"not null"`
	Commission             *decimal.Decimal `gorm:"not null"`
	Time                   int64            `gorm:"not null"`
	Comment                string           `gorm:"not null"`
	BlockID                int64            `gorm:"not null"`
	RbID                   int64            `gorm:"not null"`
}

func (dt *DltTransaction) Create(transaction *DbTransaction) error {
	return GetDB(transaction).Create(dt).Error
}
