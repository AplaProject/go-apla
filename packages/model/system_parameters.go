package model

import (
	"encoding/json"
	"strconv"
)

// SystemParameter is model
type SystemParameter struct {
	Name       string `gorm:"primary_key;not null;size:255"`
	Value      string `gorm:"not null;type:jsonb(PostgreSQL)"`
	Conditions string `gorm:"not null"`
	RbID       int64  `gorm:"not null"`
}

// TableName returns name of table
func (sp SystemParameter) TableName() string {
	return "system_parameters"
}

// Get is retrieving model from database
func (sp *SystemParameter) Get(name string) (bool, error) {
	return isFound(DBConn.Where("name = ?", name).First(sp))
}

// GetJSONField returns fields as json
func (sp *SystemParameter) GetJSONField(jsonField string, name string) (string, error) {
	var result string
	err := DBConn.Table("system_parameters").Where("name = ?", name).Select(jsonField).Row().Scan(&result)
	return result, err
}

// GetValueParameterByName returns value parameter by name
func (sp *SystemParameter) GetValueParameterByName(name, value string) (*string, error) {
	var result *string
	err := DBConn.Raw(`SELECT value->'`+value+`' FROM system_parameters WHERE name = ?`, name).Row().Scan(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetAllSystemParameters returns all system parameters
func GetAllSystemParameters() ([]SystemParameter, error) {
	parameters := new([]SystemParameter)
	if err := DBConn.Find(&parameters).Error; err != nil {
		return nil, err
	}
	return *parameters, nil
}

// ToMap is converting SystemParameter to map
func (sp *SystemParameter) ToMap() map[string]string {
	result := make(map[string]string, 0)
	result["name"] = sp.Name
	result["value"] = sp.Value
	result["conditions"] = sp.Conditions
	result["rb_id"] = strconv.FormatInt(sp.RbID, 10)
	return result
}

// SystemParameterV2 is second version model
type SystemParameterV2 struct {
	Name       string `gorm:"primary_key;not null;size:255"`
	Value      string `gorm:"not null"`
	Conditions string `gorm:"not null"`
	RbID       int64  `gorm:"not null"`
}

// TableName returns name of table
func (sp SystemParameterV2) TableName() string {
	return "system_parameters"
}

// Update is update model
func (sp SystemParameterV2) Update(value string) error {
	return DBConn.Model(sp).Where("name = ?", sp.Name).Update(`value`, value).Error
}

// SaveArray is saving array
func (sp *SystemParameterV2) SaveArray(list [][]string) error {
	ret, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return sp.Update(string(ret))
}

// Get is retrieving model from database
func (sp *SystemParameterV2) Get(name string) (bool, error) {
	return isFound(DBConn.Where("name = ?", name).First(sp))
}

// GetAllSystemParametersV2 is is retrieving all SystemParameterV2 models from database
func GetAllSystemParametersV2() ([]SystemParameterV2, error) {
	parameters := new([]SystemParameterV2)
	if err := DBConn.Find(&parameters).Error; err != nil {
		return nil, err
	}
	return *parameters, nil
}
