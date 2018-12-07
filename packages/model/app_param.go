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

package model

import "github.com/AplaProject/go-apla/packages/converter"

// AppParam is model
type AppParam struct {
	ecosystem  int64
	ID         int64  `gorm:"primary_key;not null"`
	AppID      int64  `gorm:"not null"`
	Name       string `gorm:"not null;size:100"`
	Value      string `gorm:"not null"`
	Conditions string `gorm:"not null"`
}

// TableName returns name of table
func (sp *AppParam) TableName() string {
	if sp.ecosystem == 0 {
		sp.ecosystem = 1
	}
	return `1_app_params`
}

// SetTablePrefix is setting table prefix
func (sp *AppParam) SetTablePrefix(tablePrefix string) {
	sp.ecosystem = converter.StrToInt64(tablePrefix)
}

// Get is retrieving model from database
func (sp *AppParam) Get(transaction *DbTransaction, app int64, name string) (bool, error) {
	return isFound(GetDB(transaction).Where("ecosystem=? and app_id=? and name = ?",
		sp.ecosystem, app, name).First(sp))
}

// GetAllAppParameters is returning all state parameters
func (sp *AppParam) GetAllAppParameters(app int64) ([]AppParam, error) {
	parameters := make([]AppParam, 0)
	err := DBConn.Table(sp.TableName()).Where(`ecosystem = ?`, sp.ecosystem).Find(&parameters).Error
	if err != nil {
		return nil, err
	}
	return parameters, nil
}
