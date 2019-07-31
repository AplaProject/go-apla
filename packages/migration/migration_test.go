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

package migration

import (
	"testing"
)

type dbMock struct {
	versions []string
}

func (dbm *dbMock) CurrentVersion() (string, error) {
	return dbm.versions[len(dbm.versions)-1], nil
}

func (dbm *dbMock) ApplyMigration(version, query string) error {
	dbm.versions = append(dbm.versions, version)
	return nil
}

func createDBMock(version string) *dbMock {
	return &dbMock{versions: []string{version}}
}

func TestMockMigration(t *testing.T) {
	err := migrate(createDBMock("error version"), ``, nil)
	if err.Error() != "Wrong version error version" {
		t.Error(err)
	}

	appVer := "0.0.2"

	err = migrate(createDBMock("0"), appVer, []*migration{&migration{"error version", ""}})
	if err.Error() != "Wrong version 0" {
		t.Error(err)
	}

	db := createDBMock("0.0.0")
	err = migrate(
		db, appVer,
		[]*migration{
			&migration{"0.0.1", ""},
			&migration{"0.0.2", ""},
		},
	)
	if err != nil {
		t.Error(err)
	}
	if v, _ := db.CurrentVersion(); v != "0.0.2" {
		t.Errorf("current version expected 0.0.2 get %s", v)
	}

	db = createDBMock("0.0.2")
	err = migrate(db, appVer, []*migration{
		&migration{"0.0.3", ""},
	})
	if err != nil {
		t.Error(err)
	}
	if v, _ := db.CurrentVersion(); v != "0.0.2" {
		t.Errorf("current version expected 0.0.2 get %s", v)
	}
}
