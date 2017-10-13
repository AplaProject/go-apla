package model

type TestnetKey struct {
	ID      int64  `gorm:"not null"`
	StateID int64  `gorm:"not null"`
	Private []byte `gorm:"not null;size:64"`
	Wallet  int64  `gorm:"not null"`
	Status  int32  `gorm:"not null"`
}

func (TestnetKey) TableName() string {
	return "testnet_keys"
}

func (tk *TestnetKey) GetByWallet(wallet int64) error {
	return DBConn.Where("wallet = ", wallet).First(tk).Error
}

func (tk *TestnetKey) Create() error {
	return DBConn.Create(tk).Error
}

func (tk *TestnetKey) GetGeneratedCount(ID int64, stateID int64) (int64, error) {
	count := int64(-1)
	err := DBConn.Table("testnet_keys").Where("id = ? and state_id = ?", ID, stateID).Count(&count).Error
	return count, err
}

func (tk *TestnetKey) GetAvailableCount(ID int64, stateID int64) (int64, error) {
	count := int64(-1)
	err := DBConn.Table("testnet_keys").Where("id = ? and state_id = ? and status = 0", ID, stateID).Count(&count).Error
	return count, err
}
