package model

type Pages struct {
	tableName  string
	Name       string `gorm:"primary_key;not null;size:255"`
	Value      string `gorm:not null;type:jsonb(PostgreSQL)`
	Menu       string `gorm:"not null;size:255"`
	Conditions string `gorm:not null`
	RbID       int64  `gorm:not null`
}

func (p *Pages) SetTableName(newName string) {
	p.tableName = newName
}

func (p *Pages) TableName() string {
	return p.tableName
}

func (p *Pages) GetByName(name string) error {
	return DBConn.Where("name = ?", name).Find(p).Error
}
