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
	"fmt"
	"strconv"
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/migration/updates"
	log "github.com/sirupsen/logrus"
)

const (
	eVer = `Wrong version %s`
)

var migrations = []*migration{
	// Inital migration
	&migration{"0.0.1", migrationInitial, true},

	// Initial schema
	&migration{"0.1.5", migrationInitialTables, true},
	&migration{"0.1.6", migrationInitialSchema, false},
	&migration{"0.1.7", updates.M017, true},
}

var updateMigrations = []*migration{
	&migration{"2.1.0", updates.M210, true},
	&migration{"2.2.0", updates.M220, true},
	&migration{"3.1.0", updates.M310, false},
}

type migration struct {
	version  string
	data     string
	template bool
}

type database interface {
	CurrentVersion() (string, error)
	ApplyMigration(string, string) error
}

func compareVer(a, b string) (int, error) {
	var (
		av, bv []string
		ai, bi int
		err    error
	)
	if av = strings.Split(a, `.`); len(av) != 3 {
		return 0, fmt.Errorf(eVer, a)
	}
	if bv = strings.Split(b, `.`); len(bv) != 3 {
		return 0, fmt.Errorf(eVer, b)
	}
	for i, v := range av {
		if ai, err = strconv.Atoi(v); err != nil {
			return 0, fmt.Errorf(eVer, a)
		}
		if bi, err = strconv.Atoi(bv[i]); err != nil {
			return 0, fmt.Errorf(eVer, b)
		}
		if ai < bi {
			return -1, nil
		}
		if ai > bi {
			return 1, nil
		}
	}
	return 0, nil
}

func migrate(db database, appVer string, migrations []*migration) error {
	dbVerString, err := db.CurrentVersion()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "err": err}).Errorf("parse version")
		return err
	}

	if cmp, err := compareVer(dbVerString, appVer); err != nil {
		log.WithFields(log.Fields{"type": consts.MigrationError, "err": err}).Errorf("parse version")
		return err
	} else if cmp >= 0 {
		return nil
	}

	for _, m := range migrations {
		if cmp, err := compareVer(dbVerString, m.version); err != nil {
			log.WithFields(log.Fields{"type": consts.MigrationError, "err": err}).Errorf("parse version")
			return err
		} else if cmp >= 0 {
			continue
		}
		if m.template {
			m.data, err = sqlConvert([]string{m.data})
			if err != nil {
				return err
			}
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

func runMigrations(db database, migrationList []*migration) error {
	return migrate(db, consts.VERSION, migrationList)
}

// InitMigrate applies initial migrations
func InitMigrate(db database) error {
	return runMigrations(db, migrations)
}

// UpdateMigrate applies update migrations
func UpdateMigrate(db database) error {
	return runMigrations(db, updateMigrations)
}
