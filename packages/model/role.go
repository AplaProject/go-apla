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

// Role is model
type Role struct {
	tableName   string
	ID          int64  `gorm:"primary_key;not null" json:"id"`
	DefaultPage string `gorm:"not null" json:"default_page"`
	RoleName    string `gorm:"not null" json:"role_name"`
	Deleted     int64  `gorm:"not null" json:"deleted"`
	RoleType    int64  `gorm:"not null" json:"role_type"`
}

// SetTablePrefix is setting table prefix
func (r *Role) SetTablePrefix(prefix string) {
	r.tableName = prefix + "_roles"
}

// TableName returns name of table
func (r *Role) TableName() string {
	return r.tableName
}

// Get is retrieving model from database
func (r *Role) Get(transaction *DbTransaction, id int64) (bool, error) {
	return isFound(GetDB(transaction).Where("id = ?", id).First(r))
}
