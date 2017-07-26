package model

type Menu struct {
	tableName  string
	Name       string `gorm:"primary_key;not null;size:255"`
	Value      string `gorm:not null;type:jsonb(PostgreSQL)`
	Conditions string `gorm:not null`
	RbID       int64  `gorm:not null`
}

func (m *Menu) SetTableName(newName string) {
	m.tableName = newName
}

func (m *Menu) TableName() string {
	return m.tableName
}

func (m *Menu) GetByName(name string) error {
	return DBConn.Where("name = ?", name).Find(m).Error
}
