package model

type TestnetKeys struct {
	ID      int64  `gorm:"not null"`
	StateID int64  `gorm:"not null"`
	Private []byte `gorm:"not null;size:64"`
	Wallet  int64  `gorm:"not null"`
	Status  int32  `gorm:"not null"`
}

func (tk *TestnetKeys) GetByWallet(wallet int64) error {
	return DBConn.Where("wallet = ", wallet).First(tk).Error
}

func (tk *TestnetKeys) Create() error {
	return DBConn.Create(tk).Error
}

func (tk *TestnetKeys) GetGeneratedCount(ID int64, stateID int64) (int64, error) {
	var count int64
	err := DBConn.Where("id = ? and state_id = ?", ID, stateID).Count(&count).Error
	return count, err
}

func (tk *TestnetKeys) GetAvailableCount(ID int64, stateID int64) (int64, error) {
	var count int64
	err := DBConn.Where("id = ? and state_id = ? and status = 0", ID, stateID).Count(&count).Error
	return count, err
}
