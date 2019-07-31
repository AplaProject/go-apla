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

// Signature is model
type Signature struct {
	tableName  string
	Name       string `gorm:"primary_key;not null;size:255"`
	Value      string `gorm:"not null;type:jsonb(PostgreSQL)"`
	Conditions string `gorm:"not null"`
}

// SetTablePrefix is setting table prefix
func (s *Signature) SetTablePrefix(prefix string) {
	s.tableName = prefix + "_signatures"
}

// TableName returns name of table
func (s *Signature) TableName() string {
	return s.tableName
}

// Get is retrieving model from database
func (s *Signature) Get(name string) (bool, error) {
	return isFound(DBConn.Where("name = ?", name).First(s))
}
