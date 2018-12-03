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

// StopDaemon is model
type StopDaemon struct {
	StopTime int64 `gorm:"not null"`
}

// TableName returns name of table
func (sd *StopDaemon) TableName() string {
	return "stop_daemons"
}

// Create is creating record of model
func (sd *StopDaemon) Create() error {
	return DBConn.Create(sd).Error
}

// Delete is deleting record
func (sd *StopDaemon) Delete() error {
	return DBConn.Delete(&StopDaemon{}).Error
}

// Get is retrieving model from database
func (sd *StopDaemon) Get() (bool, error) {
	return isFound(DBConn.First(sd))
}

// SetStopNow is updating daemon stopping time to now
func SetStopNow() error {
	stopTime := &StopDaemon{StopTime: time.Now().Unix()}
	return stopTime.Create()
}
