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

const BinaryTableSuffix = "_binaries"

// Binary represents record of {prefix}_binaries table
type Binary struct {
	ecosystem int64
	ID        int64
	Name      string
	Data      []byte
	Hash      string
	MimeType  string
}

// SetTablePrefix is setting table prefix
func (b *Binary) SetTablePrefix(prefix string) {
	b.ecosystem = converter.StrToInt64(prefix)
}

// SetTableName sets name of table
func (b *Binary) SetTableName(tableName string) {
	ecosystem, _ := converter.ParseName(tableName)
	b.ecosystem = ecosystem
}

// TableName returns name of table
func (b *Binary) TableName() string {
	if b.ecosystem == 0 {
		b.ecosystem = 1
	}
	return `1_binaries`
}

// Get is retrieving model from database
func (b *Binary) Get(appID int64, account, name string) (bool, error) {
	return isFound(DBConn.Where("ecosystem=? and app_id = ? AND account = ? AND name = ?",
		b.ecosystem, appID, account, name).Select("id,name,hash").First(b))
}

// Link returns link to binary data
func (b *Binary) Link() string {
	return fmt.Sprintf(`/data/%s/%d/%s/%s`, b.TableName(), b.ID, "data", b.Hash)
}

// GetByID is retrieving model from db by id
func (b *Binary) GetByID(id int64) (bool, error) {
	return isFound(DBConn.Where("id=?", id).First(b))
}
