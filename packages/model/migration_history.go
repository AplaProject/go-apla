package model

type MigrationHistory struct {
	ID          int32 `gorm:"primary_key;not null"`
	Version     int32 `gorm:"not null"`
	DateApplied int32 `gorm:"not null"`
}

func (mh *MigrationHistory) TableName() string {
	return "migration_history"
}

func (mh *MigrationHistory) Get() error {
	return DBConn.First(mh).Error
}

func (mh *MigrationHistory) Create() error {
	return DBConn.Create(mh).Error
}
