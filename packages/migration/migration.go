// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package migration

import (
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/migration/updates"
	version "github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
)

var migrations = []*migration{
	// Inital migration
	&migration{"0.0.1", migrationInitial},

	// Initial schema
	&migration{"0.1.6b9", migrationInitialSchema},

	&migration{"0.1.7", updates.M123}, // duplicate of 1.2.3 version
}

var updateMigrations = []*migration{
	&migration{"1.0.7", updates.M107},
	&migration{"1.1.4", updates.M114},
	&migration{"1.1.5", updates.M115},
	&migration{"1.2.0", updates.M120},
	&migration{"1.2.1", updates.M121},
	&migration{"1.2.2", updates.M122},
	&migration{"1.2.3", updates.M123},
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

func runMigrations(db database, migrationList []*migration) error {
	appVer, err := version.NewVersion(consts.VERSION)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.MigrationError, "err": err}).Errorf("parse version")
		return err
	}

	return migrate(db, appVer, migrationList)
}

// InitMigrate applies initial migrations
func InitMigrate(db database) error {
	return runMigrations(db, migrations)
}

// UpdateMigrate applies update migrations
func UpdateMigrate(db database) error {
	return runMigrations(db, updateMigrations)
}
