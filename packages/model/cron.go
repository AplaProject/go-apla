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
	"fmt"
)

// Cron represents record of {prefix}_cron table
type Cron struct {
	tableName string
	ID        int64
	Cron      string
	Contract  string
}

// SetTablePrefix is setting table prefix
func (c *Cron) SetTablePrefix(prefix string) {
	c.tableName = prefix + "_cron"
}

// TableName returns name of table
func (c *Cron) TableName() string {
	return c.tableName
}

// Get is retrieving model from database
func (c *Cron) Get(id int64) (bool, error) {
	return isFound(DBConn.Where("id = ?", id).First(c))
}

// GetAllCronTasks is returning all cron tasks
func (c *Cron) GetAllCronTasks() ([]*Cron, error) {
	var crons []*Cron
	err := DBConn.Table(c.TableName()).Find(&crons).Error
	return crons, err
}

// UID returns unique identifier for cron task
func (c *Cron) UID() string {
	return fmt.Sprintf("%s_%d", c.tableName, c.ID)
}
