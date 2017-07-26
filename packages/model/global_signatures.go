package model

type Signatures struct {
	tableName  string
	Name       string `gorm:"primary_key;not null;size:255"`
	Value      string `gorm:not null;type:jsonb(PostgreSQL)`
	Conditions string `gorm:not null`
	RbID       int64  `gorm:not null`
}

func (s *Signatures) TableName() string {
	return s.tableName
}

func (s *Signatures) SetTableName(tableName string) {
	s.tableName = tableName
}

func (s *Signatures) GetByName(name string) error {
	return DBConn.Where("name = ?", name).First(s).Error
}

func (s *Signatures) ExistsByName(name string) (bool, error) {
	query := DBConn.Where("name = ?", name).First(s)
	return !query.RecordNotFound(), query.Error
}
