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

// ProgressComplete status of installation progress
const ProgressComplete = "complete"

// Install is model
type Install struct {
	Progress string `gorm:"not null;size:10"`
}

// TableName returns name of table
func (i *Install) TableName() string {
	return "install"
}

// Get is retrieving model from database
func (i *Install) Get() error {
	return DBConn.Find(i).Error
}

// Create is creating record of model
func (i *Install) Create() error {
	return DBConn.Create(i).Error
}
