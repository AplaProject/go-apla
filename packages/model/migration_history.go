// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.
//
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
//
// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

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
