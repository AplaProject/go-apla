package model

// MigrationHistory is model
type MigrationHistory struct {
	Version     string `gorm:"not null"`
	DateApplied int64  `gorm:"not null"`
}

// TableName returns name of table
func (mh *MigrationHistory) TableName() string {
	return "migration_history"
}

// Get is retrieving model from database
func (mh *MigrationHistory) Get() (bool, error) {
	return isFound(DBConn.First(mh))
}

// Create is creating record of model
func (mh *MigrationHistory) Create() error {
	return DBConn.Create(mh).Error
}

func (mh *MigrationHistory) Save() error {
	return DBConn.Save(mh).Error
}
