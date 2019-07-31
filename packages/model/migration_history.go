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
	"time"
)

const noVersion = "0.0.0"

// MigrationHistory is model
type MigrationHistory struct {
	ID          int64  `gorm:"primary_key;not null"`
	Version     string `gorm:"not null"`
	DateApplied int64  `gorm:"not null"`
}

// TableName returns name of table
func (mh *MigrationHistory) TableName() string {
	return "migration_history"
}

// CurrentVersion returns current version of database migrations
func (mh *MigrationHistory) CurrentVersion() (string, error) {
	if !IsTable(mh.TableName()) {
		return noVersion, nil
	}

	err := DBConn.Last(mh).Error

	if mh.Version == "" {
		return noVersion, nil
	}

	return mh.Version, err
}

// ApplyMigration executes database schema and writes migration history
func (mh *MigrationHistory) ApplyMigration(version, query string) error {
	err := DBConn.Exec(query).Error
	if err != nil {
		return err
	}

	return DBConn.Create(&MigrationHistory{Version: version, DateApplied: time.Now().Unix()}).Error
}
