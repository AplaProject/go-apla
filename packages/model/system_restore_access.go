package model

type SystemRestoreAccess struct {
	ID        int64  `gorm:"primary_key;not_null"`
	CitizenID int64  `gorm:"not null"`
	StateID   int64  `gorm:"not null"`
	Active    int64  `gorm:"not null"`
	Time      int32  `gorm:"not null"`
	Close     int64  `gorm:"not null"`
	Secret    string `gorm:"not null"`
	RbID      int64  `gorm:"not null"`
}
