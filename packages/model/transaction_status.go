package model

type TransactionStatus struct {
	Hash      []byte `gorm:"primary_key;not null"`
	Time      int32  `gorm:"not null;"`
	Type      int32  `gorm:"not null"`
	WalletID  int64  `gorm:"not null"`
	CitizenID int64  `gorm:"not null"`
	BlockID   int64  `gorm:"not null"`
	Error     string `gorm:"not null;size 255"`
}
