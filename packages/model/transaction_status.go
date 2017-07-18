package model

type TransactionsStatus struct {
	Hash      []byte `gorm:"primary_key;not null"`
	Time      int32  `gorm:"not null;"`
	Type      int32  `gorm:"not null"`
	WalletID  int64  `gorm:"not null"`
	CitizenID int64  `gorm:"not null"`
	BlockID   int64  `gorm:"not null"`
	Error     string `gorm:"not null;size 255"`
}

func (ts *TransactionsStatus) Create() error {
	return DBConn.Create(ts).Error
}

func (ts *TransactionsStatus) Get(transactionHash []byte) error {
	return DBConn.Where("hash = ?", transactionHash).First(ts).Error
}

func (ts *TransactionsStatus) UpdateBlockID(newBlockID int64, transactionHash []byte) error {
	return DBConn.Where("hash = ?", transactionHash).Update("block_id", newBlockID).Error
}

func (ts *TransactionsStatus) SetError(errorText string, transactionHash []byte) error {
	return DBConn.Where("hash = ?", transactionHash).Update("error", errorText).Error
}
