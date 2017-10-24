package model

type MigrationHistory struct {
	ID          int32 `gorm:"primary_key;not null"`
	Version     int32 `gorm:"not null"`
	DateApplied int32 `gorm:"not null"`
}

func (mh *MigrationHistory) TableName() string {
	return "migration_history"
}

func (mh *MigrationHistory) Get() (bool, error) {
	return isFound(DBConn.First(mh))
}

func (mh *MigrationHistory) Create() error {
	return DBConn.Create(mh).Error
}
