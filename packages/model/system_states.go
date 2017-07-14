package model

type SystemStates struct {
	ID   int64 `gorm:"primary_key;not null"`
	RbID int64 `gorm:"not null"`
}
