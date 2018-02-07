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

package migration

import (
	"github.com/GenesisKernel/go-genesis/packages/consts"

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
