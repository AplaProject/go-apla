package model

import (
	"encoding/json"
	"strconv"
)

type SystemParameter struct {
	Name       string `gorm:"primary_key;not null;size:255"`
	Value      string `gorm:"not null;type:jsonb(PostgreSQL)"`
	Conditions string `gorm:"not null"`
	RbID       int64  `gorm:"not null"`
}

func (sp SystemParameter) TableName() string {
	return "system_parameters"
}

func (sp *SystemParameter) Get(name string) (bool, error) {
	return isFound(DBConn.Where("name = ?", name).First(sp))
}

func (sp *SystemParameter) GetJSONField(jsonField string, name string) (string, error) {
	var result string
	err := DBConn.Table("system_parameters").Where("name = ?", name).Select(jsonField).Row().Scan(&result)
	return result, err
}

func (sp *SystemParameter) GetValueParameterByName(name, value string) (*string, error) {
	var result *string
	err := DBConn.Raw(`SELECT value->'`+value+`' FROM system_parameters WHERE name = ?`, name).Row().Scan(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GetAllSystemParameters() ([]SystemParameter, error) {
	parameters := new([]SystemParameter)
	if err := DBConn.Find(&parameters).Error; err != nil {
		return nil, err
	}
	return *parameters, nil
}

func (sp *SystemParameter) ToMap() map[string]string {
	result := make(map[string]string, 0)
	result["name"] = sp.Name
	result["value"] = sp.Value
	result["conditions"] = sp.Conditions
	result["rb_id"] = strconv.FormatInt(sp.RbID, 10)
	return result
}

type SystemParameterV2 struct {
	Name       string `gorm:"primary_key;not null;size:255"`
	Value      string `gorm:"not null"`
	Conditions string `gorm:"not null"`
	RbID       int64  `gorm:"not null"`
}

func (sp SystemParameterV2) TableName() string {
	return "system_parameters"
}

func (sp SystemParameterV2) Update(value string) error {
	return DBConn.Model(sp).Where("name = ?", sp.Name).Update(`value`, value).Error
}

func (sp *SystemParameterV2) SaveArray(list [][]string) error {
	ret, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return sp.Update(string(ret))
}

func (sp *SystemParameterV2) Get(name string) (bool, error) {
	return isFound(DBConn.Where("name = ?", name).First(sp))
}

func GetAllSystemParametersV2() ([]SystemParameterV2, error) {
	parameters := new([]SystemParameterV2)
	if err := DBConn.Find(&parameters).Error; err != nil {
		return nil, err
	}
	return *parameters, nil
}
