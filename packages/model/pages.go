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

// Page is model
type Page struct {
	ecosystem     int64
	ID            int64  `gorm:"primary_key;not null" json:"id,omitempty"`
	Name          string `gorm:"not null" json:"name,omitempty"`
	Value         string `gorm:"not null" json:"value,omitempty"`
	Menu          string `gorm:"not null;size:255" json:"menu,omitempty"`
	ValidateCount int64  `gorm:"not null" json:"nodesCount,omitempty"`
	AppID         int64  `gorm:"column:app_id;not null" json:"app_id,omitempty"`
	Conditions    string `gorm:"not null" json:"conditions,omitempty"`
}

// SetTablePrefix is setting table prefix
func (p *Page) SetTablePrefix(prefix string) {
	p.ecosystem = converter.StrToInt64(prefix)
}

// TableName returns name of table
func (p *Page) TableName() string {
	if p.ecosystem == 0 {
		p.ecosystem = 1
	}
	return `1_pages`
}

// Get is retrieving model from database
func (p *Page) Get(name string) (bool, error) {
	return isFound(DBConn.Where("ecosystem=? and name = ?", p.ecosystem, name).First(p))
}

// Count returns count of records in table
func (p *Page) Count() (count int64, err error) {
	err = DBConn.Table(p.TableName()).Count(&count).Error
	return
}

// GetByApp returns all pages belonging to selected app
func (p *Page) GetByApp(appID int64) ([]Page, error) {
	var result []Page
	err := DBConn.Select("id, name").Where("app_id = ?", appID).Find(&result).Error
	return result, err
}
