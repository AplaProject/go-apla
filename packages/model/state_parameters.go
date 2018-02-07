// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package model

// StateParameter is model
type StateParameter struct {
	tableName  string
	ID         int64  `gorm:"primary_key;not null"`
	Name       string `gorm:"not null;size:100"`
	Value      string `gorm:"not null"`
	Conditions string `gorm:"not null"`
}

// TableName returns name of table
func (sp *StateParameter) TableName() string {
	return sp.tableName
}

// SetTablePrefix is setting table prefix
func (sp *StateParameter) SetTablePrefix(tablePrefix string) {
	sp.tableName = tablePrefix + "_parameters"
}

// Get is retrieving model from database
func (sp *StateParameter) Get(transaction *DbTransaction, name string) (bool, error) {
	return isFound(GetDB(transaction).Where("name = ?", name).First(sp))
}

// GetAllStateParameters is returning all state parameters
func (sp *StateParameter) GetAllStateParameters() ([]StateParameter, error) {
	parameters := make([]StateParameter, 0)
	err := DBConn.Table(sp.TableName()).Find(&parameters).Error
	if err != nil {
		return nil, err
	}
	return parameters, nil
}
