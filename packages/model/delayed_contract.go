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

const tableDelayedContracts = "1_delayed_contracts"

// DelayedContract represents record of 1_delayed_contracts table
type DelayedContract struct {
	ID         int64  `gorm:"primary_key;not null"`
	Contract   string `gorm:"not null"`
	KeyID      int64  `gorm:"not null"`
	EveryBlock int64  `gorm:"not null"`
	BlockID    int64  `gorm:"not null"`
	Counter    int64  `gorm:"not null"`
	Limit      int64  `gorm:"not null"`
	Delete     bool   `gorm:"not null"`
	Conditions string `gorm:"not null"`
}

// TableName returns name of table
func (DelayedContract) TableName() string {
	return tableDelayedContracts
}

// GetAllDelayedContractsForBlockID returns contracts that want to execute for blockID
func GetAllDelayedContractsForBlockID(blockID int64) ([]*DelayedContract, error) {
	var contracts []*DelayedContract
	if err := DBConn.Where("block_id = ?", blockID).Find(&contracts).Error; err != nil {
		return nil, err
	}
	return contracts, nil
}
