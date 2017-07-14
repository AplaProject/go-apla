package model

type GlobalSmartContracts struct {
	ID         int64  `gorm:"primary_key;not null"`
	Name       strign `gorm:"not null;size:100"`
	Value      []byte `gorm:not null`
	WalletID   int64  `gorm:not null`
	Active     string `gorm:not null;size:1`
	Conditions string `gorm:not null`
	Variables  []byte `gorm:not null`
	RbID       int64  `gorm:not null`
}
