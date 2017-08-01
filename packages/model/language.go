package model

type Language struct {
	tableName  string
	Name       string `gorm:"primary_key;not null;size:100"`
	Res        string `gorm:"type:jsonb(PostgreSQL)"`
	Conditions string `gorm:"not null"`
	RbID       int64  `gotm:"not null"`
}

func (l *Language) SetTableName(prefix string) {
	l.tableName = prefix + "_languages"
}

func (l *Language) TableName() string {
	return l.tableName
}

func (l *Language) Get(name string) error {
	return DBConn.Where("name = ?", name).First(l).Error
}

func (l *Language) GetAll(prefix string) ([]Language, error) {
	var result []Language
	err := DBConn.Table(prefix + "_languages").Order("name").Find(result).Error
	return result, err
}
