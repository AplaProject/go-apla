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

// SystemState is model
type SystemState struct {
	ID int64 `gorm:"primary_key;not null"`
}

// TableName returns name of table
func (ss *SystemState) TableName() string {
	return "system_states"
}

// GetAllSystemStatesIDs is retrieving all system states ids
func GetAllSystemStatesIDs() ([]int64, error) {
	states := new([]SystemState)
	if err := DBConn.Find(&states).Order("id").Error; err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(*states))
	for _, s := range *states {
		ids = append(ids, s.ID)
	}
	return ids, nil
}

// Delete is deleting record
func (ss *SystemState) Delete(transaction *DbTransaction) error {
	return GetDB(transaction).Delete(ss).Error
}
