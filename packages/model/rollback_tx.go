package model

type RollbackTx struct {
	ID        int64  `gorm:"primary_key;not null"`
	BlockID   int64  `gorm:"not null"`
	TxHash    []byte `gorm:"not null"`
	TableName string `gorm:"not null;size:255"`
	TableID   string `gorm:"not null;size:255"`
}
