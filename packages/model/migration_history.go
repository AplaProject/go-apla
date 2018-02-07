// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
