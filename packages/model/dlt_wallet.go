package model

type DltWallet struct {
	WalletID           int64  `gorm:"primary_key;not null"`
	Amount             string `gorm:"not null"`
	PublicKey          []byte `gorm:"column:public_key_0;not null"`
	NodePublicKey      []byte `gorm:"not null"`
	LastForgingDataUpd int64  `gorm:"not null"`
	Host               string `gorm:"not null"`
	AddressVote        string `gorm:"not null"`
	FuelRate           int64  `gorm:"not null"`
	SpendingContract   string `gorm:"not null"`
	ConditionsChange   string `gorm:"not null"`
	RollbackID         int64  `gorm:"not null;column:rb_id"`
}

func (DltWallet) TableName() string {
	return "dlt_wallets"
}

func (w *DltWallet) GetWalletTransaction(transaction *DbTransaction, walletID int64) (bool, error) {
	return isFound(GetDB(transaction).Where("wallet_id = ?", walletID).First(&w))
}

func (w *DltWallet) Get(walletID int64) (bool, error) {
	return isFound(DBConn.Where("wallet_id = ?", walletID).First(&w))
}
