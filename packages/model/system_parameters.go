//MIT License
//
//Copyright (c) 2016 GenesisKernel
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.
package model

import (
	"encoding/json"
)

// SystemParameter is model
type SystemParameter struct {
	ID         int64  `gorm:"primary_key;not null;"`
	Name       string `gorm:"not null;size:255"`
	Value      string `gorm:"not null"`
	Conditions string `gorm:"not null"`
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
func GetAllSystemParameters(transaction *DbTransaction) ([]SystemParameter, error) {
	parameters := new([]SystemParameter)
	if err := GetDB(transaction).Find(&parameters).Error; err != nil {
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
	return result
}

// Update is update model
func (sp SystemParameter) Update(value string) error {
	return DBConn.Model(sp).Where("name = ?", sp.Name).Update(`value`, value).Error
}

// SaveArray is saving array
func (sp *SystemParameter) SaveArray(list [][]string) error {
	ret, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return sp.Update(string(ret))
}
