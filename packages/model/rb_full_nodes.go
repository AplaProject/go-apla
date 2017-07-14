package model

type RbFullNodes struct {
	RbID                int64  `gorm:"primary_key;not null"`
	FullNodesWalletJson []byte `gorm:"not_null"`
	BlockID             int64  `gorm:"primary_key;not null"`
	PrevRbID            int64  `gorm:"not null"`
}
