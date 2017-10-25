package model

type Menu struct {
	tableName  string
	Name       string `gorm:"primary_key;not null;size:255"`
	Title      string `gorm:"not null"`
	Value      string `gorm:"not null"`
	Conditions string `gorm:"not null"`
	RbID       int64  `gorm:"not null"`
}

func (m *Menu) SetTablePrefix(prefix string) {
	m.tableName = prefix + "_menu"
}

func (m Menu) TableName() string {
	return m.tableName
}

func (m *Menu) Get(name string) (bool, error) {
	return isFound(DBConn.Where("name = ?", name).First(m))
}
