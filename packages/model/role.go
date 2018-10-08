package model

// Role is model
type Role struct {
	tableName   string
	ID          int64  `gorm:"primary_key;not null" json:"id"`
	DefaultPage string `gorm:"not null" json:"default_page"`
	RoleName    string `gorm:"not null" json:"role_name"`
	Deleted     int64  `gorm:"not null" json:"deleted"`
	RoleType    int64  `gorm:"not null" json:"role_type"`
}

// SetTablePrefix is setting table prefix
func (r *Role) SetTablePrefix(prefix string) {
	r.tableName = prefix + "_roles"
}

// TableName returns name of table
func (r *Role) TableName() string {
	return r.tableName
}

// Get is retrieving model from database
func (r *Role) Get(transaction *DbTransaction, id int64) (bool, error) {
	return isFound(GetDB(transaction).Where("id = ?", id).First(r))
}
