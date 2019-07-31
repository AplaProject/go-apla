// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package model

import (
	"github.com/AplaProject/go-apla/packages/converter"
)

// Language is model
type Language struct {
	ecosystem  int64
	ID         int64  `gorm:"primary_key;not null"`
	Name       string `gorm:"not null;size:100"`
	Res        string `gorm:"type:jsonb(PostgreSQL)"`
	Conditions string `gorm:"not null"`
}

// SetTablePrefix is setting table prefix
func (l *Language) SetTablePrefix(prefix string) {
	l.ecosystem = converter.StrToInt64(prefix)
}

// TableName returns name of table
func (l *Language) TableName() string {
	if l.ecosystem == 0 {
		l.ecosystem = 1
	}
	return `1_languages`
}

// GetAll is retrieving all records from database
func (l *Language) GetAll(prefix string) ([]Language, error) {
	result := new([]Language)
	err := DBConn.Table("1_languages").Where("ecosystem = ?", prefix).Order("name").Find(&result).Error
	return *result, err
}

// ToMap is converting model to map
func (l *Language) ToMap() map[string]string {
	result := make(map[string]string, 0)
	result["name"] = l.Name
	result["res"] = l.Res
	result["conditions"] = l.Conditions
	return result
}
