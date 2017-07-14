package model

type Rollback struct {
	ID      int64  `gorm:primary_key;not null`
	BlockID int64  `gorm:"not null"`
	Data    string `gorm:"not null"`
}
