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

// Confirmation is model
type Confirmation struct {
	BlockID int64 `gorm:"primary_key"`
	Good    int32 `gorm:"not null"`
	Bad     int32 `gorm:"not null"`
	Time    int32 `gorm:"not null"`
}

// GetGoodBlock returns last good block
func (c *Confirmation) GetGoodBlock(goodCount int) (bool, error) {
	return isFound(DBConn.Where("good >= ?", goodCount).Last(&c))
}

// GetConfirmation returns if block with blockID exists
func (c *Confirmation) GetConfirmation(blockID int64) (bool, error) {
	return isFound(DBConn.Where("block_id= ?", blockID).First(&c))
}

// Save is saving model
func (c *Confirmation) Save() error {
	return DBConn.Save(c).Error
}
