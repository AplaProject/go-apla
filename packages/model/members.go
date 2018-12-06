package model

import "github.com/AplaProject/go-apla/packages/converter"

// Member represents a ecosystem member
type Member struct {
	ecosystem  int64
	ID         int64  `gorm:"primary_key;not null"`
	MemberName string `gorm:"not null"`
	ImageID    *int64
	MemberInfo string `gorm:"type:jsonb(PostgreSQL)"`
}

// SetTablePrefix is setting table prefix
func (m *Member) SetTablePrefix(prefix string) {
	m.ecosystem = converter.StrToInt64(prefix)
}

// TableName returns name of table
func (m *Member) TableName() string {
	if m.ecosystem == 0 {
		m.ecosystem = 1
	}
	return `1_members`
}

// Count returns count of records in table
func (m *Member) Count() (count int64, err error) {
	err = DBConn.Table(m.TableName()).Where(`ecosystem=?`, m.ecosystem).Count(&count).Error
	return
}

// Get init m as member with ID
func (m *Member) Get(id int64) (bool, error) {
	return isFound(DBConn.Where("ecosystem=? and id = ?", m.ecosystem, id).First(m))
}
