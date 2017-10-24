package model

import (
	"strconv"
)

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

func GetWallets(startWalletID int64, walletsCount int) ([]DltWallet, error) {
	wallets := new([]DltWallet)
	err := DBConn.Limit(walletsCount).Where("wallet_id >= ?", startWalletID).Find(&wallets).Error
	if err != nil {
		return nil, err
	}
	return *wallets, nil
}

func (w *DltWallet) IsExistsByPublicKey(pubkey []byte) (bool, error) {
	return isFound(DBConn.Where("public_key_0 = ?", pubkey).First(w))
}
/*
func (w *DltWallet) Create() error {
	return DBConn.Create(w).Error
}
*/
func (w *DltWallet) Create(transaction *DbTransaction) error {
	return GetDB(transaction).Create(w).Error
}

func (w *DltWallet) GetVotes(limit int) ([]map[string]string, error) {
	result := make([]map[string]string, 0)

	wallets := make([]DltWallet, 0)
	err := DBConn.
		Select([]string{"address_vote", "sum(amount) as sum"}).
		Where("address_vote != ''").
		Group("address_vote").
		Order("sum(amount) desc").
		Limit(limit).
		Find(&wallets).Error
	if err != nil {
		return nil, err
	}

	for _, wallet := range wallets {
		newRow := make(map[string]string)
		newRow["amount"] = wallet.Amount
		newRow["address_vote"] = wallet.AddressVote
		result = append(result, newRow)
	}
	return result, nil
}

func (w *DltWallet) ToMap() map[string]string {
	result := make(map[string]string, 0)
	result["wallet_id"] = strconv.FormatInt(w.WalletID, 10)
	result["amount"] = w.Amount
	result["public_key_0"] = string(w.PublicKey)
	result["node_public_key"] = string(w.NodePublicKey)
	result["last_forgind_data_upd"] = strconv.FormatInt(w.LastForgingDataUpd, 10)
	result["host"] = w.Host
	result["address_vote"] = w.AddressVote
	result["fuel_rate"] = strconv.FormatInt(w.FuelRate, 10)
	result["spending_contract"] = w.SpendingContract
	result["conditions_change"] = w.ConditionsChange
	result["rb_id"] = strconv.FormatInt(w.RollbackID, 10)
	return result
}

func (w *DltWallet) GetNewFuelRate() error {
	return DBConn.Table("dlt_wallets").Where("fuel_rate !=0").Select("fuel_rate").Group("fuel_rate").Order("sum(amount)").Limit(1).Find(w).Error
}

func (w *DltWallet) GetAddressVotes() ([]string, error) {
	rows, err := DBConn.Raw(`SELECT address_vote FROM dlt_wallets WHERE address_vote !='' AND amount > 10000000000000000000000 GROUP BY address_vote ORDER BY sum(amount) DESC LIMIT 100`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	addresses := make([]string, 0)
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

func WalletCreateTable() error {
	return DBConn.CreateTable(&DltWallet{}).Error
}
