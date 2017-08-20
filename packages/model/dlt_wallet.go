package model

import (
	"strconv"

	"github.com/jinzhu/gorm"
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

func (w *DltWallet) GetWallet(walletID int64) error {
	return handleError(DBConn.Where("wallet_id = ?", walletID).First(&w).Error)
}

func GetWallets(startWalletID int64, walletsCount int) ([]DltWallet, error) {
	wallets := new([]DltWallet)
	err := DBConn.Limit(walletsCount).Where("wallet_id >= ?", startWalletID).Find(wallets).Error
	if err != nil {
		return nil, err
	}
	return *wallets, nil
}

func (w *DltWallet) IsExistsByPublicKey(pubkey []byte) (bool, error) {
	query := DBConn.Where("public_key_0 = ?", pubkey).First(w)
	if query.Error == gorm.ErrRecordNotFound {
		return false, nil
	}
	return !query.RecordNotFound(), query.Error
}

func (w *DltWallet) IsExists() (bool, error) {
	query := DBConn.Where("wallet_id = ?", w.WalletID).First(w)
	if query.Error == gorm.ErrRecordNotFound {
		return false, nil
	}
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
	return DBConn.Table("dlt_wallets").Where("fuel_rate !=0").Select("fuel_rate").Group("fuel_rate").Order("sum(amount)").First(w).Error
}

func (w *DltWallet) GetAddressVotes() ([]string, error) {
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

func WalletCreateTable() error {
	return DBConn.CreateTable(&DltWallet{}).Error
}
