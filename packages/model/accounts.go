package model

import "github.com/shopspring/decimal"

type Account struct {
	tableName string
	ID        int64           `gorm:"primary_key;not null"`
	Amount    decimal.Decimal `gorm:"not null"`
	Onhold    int64           `gorm:"not null"`
	AgencyID  int64           `gorm:"not null"`
	CitizenID int64           `gorm:"not null"`
	CompanyID int64           `gorm:"not null"`
	RbID      int64           `gorm:"not null"`
}

func (a *Account) SetTablePrefix(prefix int64) {
	a.tableName = string(prefix) + "_accounts"
}

func (a *Account) TableName() string {
	return a.tableName
}

func (a *Account) Get(accountID int64) (bool, error) {
	return isFound(DBConn.Where("id = ?", accountID).First(a))
}
