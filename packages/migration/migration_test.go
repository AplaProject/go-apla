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
