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

// InfoBlock is model
type InfoBlock struct {
	Hash           []byte `gorm:"not null"`
	EcosystemID    int64  `gorm:"not null default 0"`
	KeyID          int64  `gorm:"not null default 0"`
	NodePosition   string `gorm:"not null default 0"`
	BlockID        int64  `gorm:"not null"`
	Time           int64  `gorm:"not null"`
	CurrentVersion string `gorm:"not null"`
	Sent           int8   `gorm:"not null"`
	RollbacksHash  []byte `gorm:"not null"`
}

// TableName returns name of table
func (ib *InfoBlock) TableName() string {
	return "info_block"
}

// Get is retrieving model from database
func (ib *InfoBlock) Get() (bool, error) {
	return isFound(DBConn.Last(ib))
}

// Update is update model
func (ib *InfoBlock) Update(transaction *DbTransaction) error {
	return GetDB(transaction).Model(&InfoBlock{}).Updates(ib).Error
}

// GetUnsent is retrieving model from database
func (ib *InfoBlock) GetUnsent() (bool, error) {
	return isFound(DBConn.Where("sent = ?", "0").First(&ib))
}

// Create is creating record of model
func (ib *InfoBlock) Create(transaction *DbTransaction) error {
	return GetDB(transaction).Create(ib).Error
}

// MarkSent update model sent field
func (ib *InfoBlock) MarkSent() error {
	return DBConn.Model(ib).Update("sent", 1).Error
}

// UpdRollbackHash update model rollbacks_hash field
func UpdRollbackHash(transaction *DbTransaction, hash []byte) error {
	return GetDB(transaction).Model(&InfoBlock{}).Update("rollbacks_hash", hash).Error
}

// BlockGetUnsent returns InfoBlock
func BlockGetUnsent() (*InfoBlock, error) {
	ib := &InfoBlock{}
	found, err := ib.GetUnsent()
	if !found {
		return nil, err
	}
	return ib, err
}

// Marshall returns block as []byte
func (ib *InfoBlock) Marshall() []byte {
	if ib != nil {
		toBeSent := converter.DecToBin(ib.BlockID, 3)
		return append(toBeSent, ib.Hash...)
	}
	return []byte{}
}
