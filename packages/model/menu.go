package model

// Menu is model
type Menu struct {
	tableName  string
	Name       string `gorm:"primary_key;not null;size:255"`
	Title      string `gorm:"not null"`
	Value      string `gorm:"not null"`
	Conditions string `gorm:"not null"`
	RbID       int64  `gorm:"not null"`
}

// SetTablePrefix is setting table prefix
func (m *Menu) SetTablePrefix(prefix string) {
	m.tableName = prefix + "_menu"
}

// TableName returns name of table
func (m Menu) TableName() string {
	return m.tableName
}

// Get is retrieving model from database
func (m *Menu) Get(name string) (bool, error) {
	return isFound(DBConn.Where("name = ?", name).First(m))
}
