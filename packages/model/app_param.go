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

import "github.com/AplaProject/go-apla/packages/converter"

// AppParam is model
type AppParam struct {
	ecosystem  int64
	ID         int64  `gorm:"primary_key;not null"`
	AppID      int64  `gorm:"not null"`
	Name       string `gorm:"not null;size:100"`
	Value      string `gorm:"not null"`
	Conditions string `gorm:"not null"`
}

// TableName returns name of table
func (sp *AppParam) TableName() string {
	if sp.ecosystem == 0 {
		sp.ecosystem = 1
	}
	return `1_app_params`
}

// SetTablePrefix is setting table prefix
func (sp *AppParam) SetTablePrefix(tablePrefix string) {
	sp.ecosystem = converter.StrToInt64(tablePrefix)
}

// Get is retrieving model from database
func (sp *AppParam) Get(transaction *DbTransaction, app int64, name string) (bool, error) {
	return isFound(GetDB(transaction).Where("ecosystem=? and app_id=? and name = ?",
		sp.ecosystem, app, name).First(sp))
}

// GetAllAppParameters is returning all state parameters
func (sp *AppParam) GetAllAppParameters(app int64) ([]AppParam, error) {
	parameters := make([]AppParam, 0)
	err := DBConn.Table(sp.TableName()).Where(`ecosystem = ?`, sp.ecosystem).Find(&parameters).Error
	if err != nil {
		return nil, err
	}
	return parameters, nil
}
