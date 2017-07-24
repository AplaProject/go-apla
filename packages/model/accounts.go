package model

import "github.com/shopspring/decimal"

type Accounts struct {
	tableName string
	ID        int64           `gorm:"primary_key;not null"`
	Amount    decimal.Decimal `gorm:"not null"`
	Onhold    int64           `gorm:"not null"`
	AgencyID  int64           `gorm:"not null"`
	CitizenID int64           `gorm:"not null"`
	CompanyID int64           `gorm:"not null"`
	RbID      int64           `gorm:"not null"`
}

func (a *Accounts) SetTablePrefix(prefix int64) {
	a.tableName = string(prefix) + "_accounts"
}

func (a *Accounts) TableName() string {
	return a.tableName
}

func (a *Accounts) Get(accountID int64) error {
	return DBConn.Where("id = ?", accountID).First(a).Error
}
