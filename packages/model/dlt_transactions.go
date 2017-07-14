package model

import "github.com/shopspring/decimal"

type DltTransactions struct {
	ID                     int64           `gorm:"primary_key;not null"`
	SenderWalletID         int64           `gorm:"not null"`
	RecepientWalletID      int64           `gorm:"not null"`
	RecepientWalletAddress string          `gorm:"not null;size:32"`
	Amount                 decimal.Decimal `gorm:"not null"`
	Comission              decimal.Decimal `gorm:"not null"`
	Time                   int32           `gorm:"not null"`
	Comment                string          `gorm:"not null"`
	BlockID                int64           `gorm:"not null"`
	RbID                   int64           `gorm:"not null"`
}
