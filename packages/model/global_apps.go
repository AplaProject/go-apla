package model

type GlobalApps struct {
	Name   string `gorm:"private_key;not null;size:100"`
	Done   int32  `gorm:"not null"`
	Blocks string `gorm:"not null"`
}
