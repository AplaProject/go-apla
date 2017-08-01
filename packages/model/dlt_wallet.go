package model

import "github.com/shopspring/decimal"

type DltWallet struct {
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
	RollbackID         int64           `gorm:"not null;column:rb_id"`
}

func (w *DltWallet) GetWallet(walletID int64) error {
	if err := DBConn.Where("wallet_id = ", walletID).First(&w).Error; err != nil {
		return err
	}
	return nil
}

func GetWallets(startWalletID int64, walletsCount int) ([]DltWallet, error) {
	wallets := new([]DltWallet)
	err := DBConn.Limit(walletsCount).Where("wallet_id >= ?", startWalletID).Find(wallets).Error
	if err != nil {
		return nil, err
	}
	return *wallets, nil
}

func (w *DltWallet) IsExistsByPublicKey() (bool, error) {
	query := DBConn.Where("public_key_0 = ", w.PublicKey).First(w)
	return !query.RecordNotFound(), query.Error
}

func (w *DltWallet) IsExists() (bool, error) {
	query := DBConn.Where("wallet_id = ", w.WalletID).First(w)
	return !query.RecordNotFound(), query.Error
}

func (w *DltWallet) Create() error {
	return DBConn.Create(w).Error
}

func (w *DltWallet) GetVotes(limit int) ([]map[string]string, error) {
	result := make([]map[string]string, 0)

	var wallets []DltWallet
	err := DBConn.
		Select([]string{"address_vote", "sum(amount) as sum"}).
		Where("address_vote != ''").
		Group("address_vote").
		Order("sum(amount) desc").
		Limit(limit).
		Find(wallets).Error
	if err != nil {
		return nil, err
	}

	for _, wallet := range wallets {
		newRow := make(map[string]string)
		newRow["amount"] = wallet.Amount.String()
		newRow["address_vote"] = wallet.AddressVote
		result = append(result, newRow)
	}
	return result, nil
}

func (w *DltWallet) ToMap() map[string]string {
	result := make(map[string]string, 0)
	result["wallet_id"] = string(w.WalletID)
	result["amount"] = w.Amount.String()
	result["public_key_0"] = string(w.PublicKey)
	result["node_public_key"] = string(w.NodePublicKey)
	result["last_forgind_data_upd"] = string(w.LastForgingDataUpd)
	result["host"] = w.Host
	result["address_vote"] = w.AddressVote
	result["fuel_rate"] = string(w.FuelRate)
	result["spending_contract"] = w.SpendingContract
	result["conditions_change"] = w.ConditionsChange
	result["rb_id"] = string(w.RollbackID)
	return result
}
