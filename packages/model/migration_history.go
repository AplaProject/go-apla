package model

type MigrationHistory struct {
	ID          int32 `gorm:"primary_key;not null"`
	Version     int32 `gorm:"not null"`
	DateApplied int32 `gorm:"not null"`
}
