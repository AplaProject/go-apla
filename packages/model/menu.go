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

// Menu is model
type Menu struct {
	ecosystem  int64
	ID         int64  `gorm:"primary_key;not null" json:"id"`
	Name       string `gorm:"not null" json:"name"`
	Title      string `gorm:"not null" json:"title"`
	Value      string `gorm:"not null" json:"value"`
	Conditions string `gorm:"not null" json:"conditions"`
}

// SetTablePrefix is setting table prefix
func (m *Menu) SetTablePrefix(prefix string) {
	m.ecosystem = converter.StrToInt64(prefix)
}

// TableName returns name of table
func (m Menu) TableName() string {
	if m.ecosystem == 0 {
		m.ecosystem = 1
	}
	return `1_menu`
}

// Get is retrieving model from database
func (m *Menu) Get(name string) (bool, error) {
	return isFound(DBConn.Where("ecosystem=? and name = ?", m.ecosystem, name).First(m))
}
