package model

// ProgressComplete status of installation progress
const ProgressComplete = "complete"

// Install is model
type Install struct {
	Progress string `gorm:"not null;size:10"`
}

// TableName returns name of table
func (i *Install) TableName() string {
	return "install"
}

// Get is retrieving model from database
func (i *Install) Get() error {
	return DBConn.Find(i).Error
}

// Create is creating record of model
func (i *Install) Create() error {
	return DBConn.Create(i).Error
}
