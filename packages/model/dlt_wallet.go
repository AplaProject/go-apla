package model

//import "github.com/shopspring/decimal"

type Wallet struct {
	WalletID           int64           `gorm:"primary_key;not null"`
	//Amount             decimal.Decimal `gorm:"not null"`
	Amount				int64		   `gorm:"not null"`
	PublicKey          []byte          `gorm:"column:public_key_0;not null"`
	NodePublicKey      []byte          `gorm:"not null"`
	LastForgingDataUpd int64           `gorm:"not null default 0"`
	Host               string          `gorm:"not null default ''"`
	AddressVote        string          `gorm:"not null default ''"`
	FuelRate           int64           `gorm:"not null default 0"`
	SpendingContract   string          `gorm:"not null default ''"`
	ConditionsChange   string          `gorm:"not null default ''"`
	RollbackID         int64           `gorm:"not null default 0"`
}

func (Wallet) TableName() string {
	return "dlt_wallets"
}

func (w *Wallet) GetWallet(walletID int64) error {
	if err := DBConn.Where("wallet_id = ", walletID).First(&w).Error; err != nil {
		return err
	}
	return nil
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

func WalletCreateTable() error {
	return DBConn.CreateTable(&Wallet{}).Error
}