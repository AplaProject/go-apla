// MIT License
//
// Copyright (c) 2016 GenesisKernel
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
package migration

import (
	"testing"

	version "github.com/hashicorp/go-version"
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
	err := migrate(createDBMock("error version"), nil, nil)
	if err.Error() != "Malformed version: error version" {
		t.Error(err)
	}

	appVer := version.Must(version.NewVersion("0.0.2"))

	err = migrate(createDBMock("0"), appVer, []*migration{&migration{"error version", ""}})
	if err.Error() != "Malformed version: error version" {
		t.Error(err)
	}

	db := createDBMock("0")
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
