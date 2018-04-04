package model

// Member represents a ecosystem member
type Member struct {
	tableName  string
	ID         int64  `gorm:"primary_key;not null"`
	MemberName string `gorm:"not null"`
	Avatar     string `gorm:"not null"`
}

// SetTablePrefix is setting table prefix
func (m *Member) SetTablePrefix(prefix string) {
	m.tableName = prefix + "_members"
}

// TableName returns name of table
func (m *Member) TableName() string {
	return m.tableName
}

// Count returns count of member
func (m *Member) Count() (count int64, err error) {
	err = DBConn.Table(m.TableName()).Count(&count).Error
	return
}

// Get init m as member with ID
func (m *Member) Get(ID int64) error {
	return DBConn.First(m, ID).Error
}
