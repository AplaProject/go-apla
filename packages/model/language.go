package model

import "strconv"

// Language is model
type Language struct {
	tableName  string
	Name       string `gorm:"primary_key;not null;size:100"`
	Res        string `gorm:"type:jsonb(PostgreSQL)"`
	Conditions string `gorm:"not null"`
	RbID       int64  `gotm:"not null"`
}

func (l *Language) SetTablePrefix(tablePrefix string) {
	l.tableName = tablePrefix + "_languages"
}

// TableName returns name of table
func (l *Language) TableName() string {
	return l.tableName
}

// Get is retrieving all records from database
func (l *Language) GetAll(prefix string) ([]Language, error) {
	result := new([]Language)
	err := DBConn.Table(prefix + "_languages").Order("name").Find(&result).Error
	return *result, err
}

func (l *Language) ToMap() map[string]string {
	result := make(map[string]string, 0)
	result["name"] = l.Name
	result["res"] = l.Res
	result["conditions"] = l.Conditions
	result["rb_id"] = strconv.FormatInt(l.RbID, 10)
	return result
}
