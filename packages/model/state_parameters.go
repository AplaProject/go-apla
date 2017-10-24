package model

import (
	"strconv"
)

type StateParameter struct {
	tableName  string
	Name       string `gorm:"primary_key;not null;size:100"`
	Value      string `gorm:"not null"`
	Conditions string `gorm:"not null"`
	RbID       int64  `gorm:"not null"`
}

func (sp *StateParameter) TableName() string {
	return sp.tableName
}

func (sp *StateParameter) SetTablePrefix(tablePrefix string) {
	sp.tableName = tablePrefix + "_parameters"
}

func (sp *StateParameter) GetByNameTransaction(transaction *DbTransaction, name string) (bool, error) {
	return isFound(GetDB(transaction).Where("name = ?", name).First(sp))
}

func (sp *StateParameter) GetByName(name string) (bool, error) {
	return isFound(DBConn.Where("name = ?", name).First(sp))
}

func (sp *StateParameter) GetAllStateParameters() ([]StateParameter, error) {
	parameters := make([]StateParameter, 0)
	err := DBConn.Table(sp.TableName()).Find(&parameters).Error
	if err != nil {
		return nil, err
	}
	return parameters, nil
}

func (sp *StateParameter) ToMap() map[string]string {
	result := make(map[string]string, 0)
	result["name"] = sp.Name
	result["value"] = sp.Value
	result["conditions"] = sp.Conditions
	result["rb_id"] = strconv.FormatInt(sp.RbID, 10)
	return result
}
