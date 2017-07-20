package model

import "github.com/shopspring/decimal"

type Wallet struct {
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

func (w *Wallet) GetWallet(walletID int64) error {
	return DBConn.Where("wallet_id = ", walletID).First(&w).Error
}

func GetWallets(startWalletID int64, walletsCount int) ([]Wallet, error) {
	wallets := new([]Wallet)
	err := DBConn.Limit(walletsCount).Where("wallet_id >= ?", startWalletID).Find(wallets).Error
	if err != nil {
		return nil, err
	}
	return *wallets, nil
}

func (w *Wallet) IsExistsByPublicKey() (bool, error) {
	query := DBConn.Where("public_key_0 = ", w.PublicKey).First(w)
	return !query.RecordNotFound(), query.Error
}

func (w *Wallet) IsExists() (bool, error) {
	query := DBConn.Where("wallet_id = ", w.WalletID).First(w)
	return !query.RecordNotFound(), query.Error
}

func (w *Wallet) Create() error {
	return DBConn.Create(w).Error
}

/*
func (db *DCDB) GetVotes() ([]map[string]string, error) {
	return db.GetAll(`SELECT address_vote, sum(amount) as sum FROM dlt_wallets WHERE address_vote !=''
	 GROUP BY address_vote ORDER BY sum(amount) DESC LIMIT 10`, -1)
}
*/
