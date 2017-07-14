package model

type GlobalPages struct {
	Name       string `gorm:"primary_key;not null;size:255"`
	Value      string `gorm:not null;type:jsonb(PostgreSQL)`
	Menu       string `gorm:"not null;size:255"`
	Conditions string `gorm:not null`
	RbID       int64  `gorm:not null`
}
