package model

// Member represents a ecosystem member
type Member struct {
	tableName  string
	ID         int64  `gorm:"primary_key;not null"`
	MemberName string `gorm:"not null"`
	ImageID    *int64
	MemberInfo []byte
}

// SetTablePrefix is setting table prefix
func (m *Member) SetTablePrefix(prefix string) {
	m.tableName = prefix + "_members"
}

// TableName returns name of table
func (m *Member) TableName() string {
	return m.tableName
}

// Count returns count of records in table
func (m *Member) Count() (count int64, err error) {
	err = DBConn.Table(m.TableName()).Count(&count).Error
	return
}

// Get init m as member with ID
func (m *Member) Get(id int64) (bool, error) {
	return isFound(DBConn.Where("id = ?", id).First(m))
}
