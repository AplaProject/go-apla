package model

type Signatures struct {
	tableName  string
	Name       string `gorm:"primary_key;not null;size:255"`
	Value      string `gorm:not null;type:jsonb(PostgreSQL)`
	Conditions string `gorm:"not null"`
	RbID       int64  `gorm:"not null"`
}

func (s *Signatures) SetTableName(prefix string) {
	s.tableName = prefix + "_signatures"
}

func (s *Signatures) TableName() string {
	return s.tableName
}

func (s *Signatures) Get(name string) error {
	return DBConn.Where("name = ?", name).First(s).Error
}

func (s *Signatures) GetAllOredered(prefix string) ([]Signatures, error) {
	var result []Signatures
	err := DBConn.Table(prefix + "_signatures").Order("name").Find(result).Error
	return result, err
}

func (s *Signatures) ToMap() map[string]string {
	var result map[string]string
	result["name"] = s.Name
	result["value"] = s.Value
	result["conditions"] = s.Conditions
	result["rb_id"] = string(s.RbID)
	return result
}
