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
	"strings"

	"github.com/AplaProject/go-apla/packages/converter"
)

// ExternalBlockchain represents a txinfo table
type ExternalBlockchain struct {
	Id               int64  `gorm:"primary_key;not null"`
	Value            string `gorm:"not null"`
	ExternalContract string `gorm:"not null"`
	ResultContract   string `gorm:"not null"`
	Url              string `gorm:"not null"`
	Uid              string `gorm:"not null"`
	TxTime           int64  `gorm:"not null"`
	Sent             int64  `gorm:"not null"`
	Hash             []byte `gorm:"not null"`
	Attempts         int64  `gorm:"not null"`
}

// GetExternalList returns the list of network tx
func GetExternalList() (list []ExternalBlockchain, err error) {
	err = DBConn.Table("external_blockchain").
		Order("id").Scan(&list).Error
	return
}

// DelExternalList deletes sent tx
func DelExternalList(list []int64) error {
	slist := make([]string, len(list))
	for i, v := range list {
		slist[i] = converter.Int64ToStr(v)
	}
	return DBConn.Exec("delete from external_blockchain where id in (" +
		strings.Join(slist, `,`) + ")").Error
}

func HashExternalTx(id int64, hash []byte) error {
	return DBConn.Exec("update external_blockchain set hash=?, sent = 1 where id = ?", hash, id).Error
}

func IncExternalAttempt(id int64) error {
	return DBConn.Exec("update external_blockchain set attempts=attempts+1 where id = ?", id).Error
}
