package model

import "github.com/shopspring/decimal"

type Wallet struct {
	tableName          string
	WalletID           int64           `gorm:"primary_key;not null"`
	Amount             decimal.Decimal `gorm:"not null"`
	PublicKey          []byte          `gorm:"column:publick_key_0;not null"`
	NodePublicKey      []byte          `gorm:"not null"`
	LastForgingDataUpd int64           `gorm:"not null"`
	Host               string          `gorm:"not null"`
	AddressVote        string          `gorm:"not null"`
	FuelRate           int64           `gorm:"not null"`
	SpendingContract   string          `gorm:"not null"`
	ConditionsChange   string          `gorm:"not null"`
	RollbackID         int64           `gorm:"not null"`
}

func NewWallet() *Wallet {
	return &Wallet{tableName: "dlt_wallets"}
}

func (w *Wallet) TableName() string {
	return w.tableName
}

func (w *Wallet) SetTableName(tablePrefix string) {
	w.tableName = tablePrefix + "_dlt_wallets"
}

func (w *Wallet) GetWallet(walletID int64) error {
	if err := DBConn.Where("wallet_id = ", walletID).First(&w).Error; err != nil {
		return err
	}
	return nil
}
