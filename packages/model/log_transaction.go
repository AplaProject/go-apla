package model

type LogTransactions struct {
	Hash []byte `gorm:"primary_key;not null"`
	Time int32  `gorm:"not null"`
}
