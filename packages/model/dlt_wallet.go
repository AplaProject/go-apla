package model

import "github.com/shopspring/decimal"

type Wallet struct {
	WalletID           int64           `gorm:"primary_key;not null"`
	Amount             decimal.Decimal `gorm:"not null"`
	PublicKey          []byte          `gorm:"column:public_key_0;not null"`
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
	return DBConn.Where("wallet_id = ?", walletID).First(&w).Error
}

func GetWallets(startWalletID int64, walletsCount int) ([]Wallet, error) {
	wallets := new([]Wallet)
	err := DBConn.Limit(walletsCount).Where("wallet_id >= ?", startWalletID).Find(wallets).Error
	if err != nil {
		return nil, err
	}
	return *wallets, nil
}

func (w *Wallet) IsExistsByPublicKey(pubkey []byte) (bool, error) {
	query := DBConn.Where("public_key_0 = ?", pubkey).First(w)
	return !query.RecordNotFound(), query.Error
}

func (w *Wallet) IsExists() (bool, error) {
	query := DBConn.Where("wallet_id = ?", w.WalletID).First(w)
	return !query.RecordNotFound(), query.Error
}

func (w *Wallet) Create() error {
	return DBConn.Create(w).Error
}

func (w *Wallet) GetNewFuelRate() (string, error) {
	return Single(`SELECT fuel_rate FROM dlt_wallets WHERE fuel_rate !=0 GROUP BY fuel_rate ORDER BY sum(amount) DESC LIMIT 1`).String()
}

func (w *Wallet) GetAddressVotes() ([]string, error) {
	rows, err := DBConn.Raw(`SELECT address_vote FROM dlt_wallets WHERE address_vote !='' AND amount > 10000000000000000000000 GROUP BY address_vote ORDER BY sum(amount) DESC LIMIT 100`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var addresses []string
	for rows.Next() {
		var addressVote string
		err := rows.Scan(&addressVote)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, addressVote)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return addresses, nil
}
