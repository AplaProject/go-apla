package migration

import (
	"github.com/AplaProject/go-apla/packages/consts"

	version "github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
)

var migrations = []*migration{
	// Inital migration
	&migration{"0.0.1", migrationInitial},

	// Initial schema
	&migration{"0.1.6b9", migrationInitialSchema},
}

type migration struct {
	version string
	data    string
}

type database interface {
	CurrentVersion() (string, error)
	ApplyMigration(string, string) error
}

func migrate(db database, appVer *version.Version, migrations []*migration) error {
	dbVerString, err := db.CurrentVersion()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "err": err}).Errorf("parse version")
		return err
	}

	dbVer, err := version.NewVersion(dbVerString)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.MigrationError, "err": err}).Errorf("parse version")
		return err
	}

	// if the database version is up-to-date
	if !dbVer.LessThan(appVer) {
		return nil
	}

	for _, m := range migrations {
		mgrVer, err := version.NewVersion(m.version)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.MigrationError, "err": err}).Errorf("parse version")
			return err
		}
		if !dbVer.LessThan(mgrVer) {
			continue
		}

		err = db.ApplyMigration(m.version, m.data)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "err": err, "version": m.version}).Errorf("apply migration")
			return err
		}

		log.WithFields(log.Fields{"version": m.version}).Info("apply migration")
	}

	return nil
}

// Migrate applies migrations
func Migrate(db database) error {
	appVer, err := version.NewVersion(consts.VERSION)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.MigrationError, "err": err}).Errorf("parse version")
		return err
	}

	return migrate(db, appVer, migrations)
}
