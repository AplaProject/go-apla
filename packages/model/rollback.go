package model

// Rollback is model
type Rollback struct {
	RbID    int64  `gorm:"primary_key;not null"`
	BlockID int64  `gorm:"not null"`
	Data    string `gorm:"not null;type:jsonb(PostgreSQL)"`
}

// TableName returns name of table
func (Rollback) TableName() string {
	return "rollback"
}

// Get is retrieving model from database
func (r *Rollback) Get(rollbackID int64) (bool, error) {
	return isFound(DBConn.Where("rb_id = ?", rollbackID).First(r))
}

// Create is creating record of model
func (r *Rollback) Create(transaction *DbTransaction) error {
	return GetDB(transaction).Create(r).Error
}

func (r *Rollback) Delete() error {
	return DBConn.Delete(r).Error
}
