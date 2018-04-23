package migration

import (
	"testing"

	"github.com/stretchr/testify/require"

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
	require.EqualError(t, migrate(createDBMock("error version"), nil, nil), "Malformed version: error version")

	appVer := version.Must(version.NewVersion("0.0.2"))

	require.EqualError(t, migrate(createDBMock("0"), appVer, []*migration{&migration{"error version", ""}}), "Malformed version: error version")

	db := createDBMock("0")
	require.NoError(t, migrate(
		db, appVer,
		[]*migration{
			&migration{"0.0.1", ""},
			&migration{"0.0.2", ""},
		},
	))

	if v, _ := db.CurrentVersion(); v != "0.0.2" {
		t.Errorf("current version expected 0.0.2 get %s", v)
	}

	db = createDBMock("0.0.2")
	require.NoError(t, migrate(db, appVer, []*migration{
		&migration{"0.0.3", ""},
	}))

	v, _ := db.CurrentVersion()
	require.Equalf(t, "0.0.2", v, "current version expected 0.0.2 get %s", v)
}
