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
	"fmt"

	"github.com/AplaProject/go-apla/packages/converter"
)

// Key is model
type Key struct {
	ecosystem    int64
	accountKeyID int64 `gorm:"-"`

	ID        int64  `gorm:"primary_key;not null"`
	AccountID string `gorm:"column:account;not null"`
	PublicKey []byte `gorm:"column:pub;not null"`
	Amount    string `gorm:"not null"`
	Maxpay    string `gorm:"not null"`
	Deleted   int64  `gorm:"not null"`
	Blocked   int64  `gorm:"not null"`
}

// SetTablePrefix is setting table prefix
func (m *Key) SetTablePrefix(prefix int64) *Key {
	m.ecosystem = prefix
	return m
}

// TableName returns name of table
func (m Key) TableName() string {
	if m.ecosystem == 0 {
		m.ecosystem = 1
	}
	return `1_keys`
}

// Get is retrieving model from database
func (m *Key) Get(wallet int64) (bool, error) {
	return isFound(DBConn.Where("id = ? and ecosystem = ?", wallet, m.ecosystem).First(m))
}

func (m *Key) AccountKeyID() int64 {
	if m.accountKeyID == 0 {
		m.accountKeyID = converter.StringToAddress(m.AccountID)
	}
	return m.accountKeyID
}

// KeyTableName returns name of key table
func KeyTableName(prefix int64) string {
	return fmt.Sprintf("%d_keys", prefix)
}

// GetKeysCount returns common count of keys
func GetKeysCount() (int64, error) {
	var cnt int64
	row := DBConn.Raw(`SELECT count(*) key_count FROM "1_keys" WHERE ecosystem = 1`).Select("key_count").Row()
	err := row.Scan(&cnt)
	return cnt, err
}
