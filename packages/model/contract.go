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

// Contract represents record of 1_contracts table
type Contract struct {
	tableName   string
	ID          int64  `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Value       string `json:"value,omitempty"`
	WalletID    int64  `json:"wallet_id,omitempty"`
	Active      bool   `json:"active,omitempty"`
	TokenID     int64  `json:"token_id,omitempty"`
	Conditions  string `json:"conditions,omitempty"`
	AppID       int64  `json:"app_id,omitempty"`
	EcosystemID int64  `gorm:"column:ecosystem" json:"ecosystem_id,omitempty"`
}

// TableName returns name of table
func (c *Contract) TableName() string {
	return `1_contracts`
}

// GetList is retrieving records from database
func (c *Contract) GetList(offset, limit int64) ([]Contract, error) {
	result := new([]Contract)
	err := DBConn.Table(c.TableName()).Offset(offset).Limit(limit).Order("id asc").Find(&result).Error
	return *result, err
}

// GetFromEcosystem retrieving ecosystem contracts from database
func (c *Contract) GetFromEcosystem(db *DbTransaction, ecosystem int64) ([]Contract, error) {
	result := new([]Contract)
	err := GetDB(db).Table(c.TableName()).Where("ecosystem = ?", ecosystem).Order("id asc").Find(&result).Error
	return *result, err
}

// Count returns count of records in table
func (c *Contract) Count() (count int64, err error) {
	err = DBConn.Table(c.TableName()).Count(&count).Error
	return
}

func (c *Contract) GetListByEcosystem(offset, limit int64) ([]Contract, error) {
	var list []Contract
	err := DBConn.Table(c.TableName()).Offset(offset).Limit(limit).
		Order("id").Where("ecosystem = ?", c.EcosystemID).
		Find(&list).Error
	return list, err
}

func (c *Contract) CountByEcosystem() (n int64, err error) {
	err = DBConn.Table(c.TableName()).Where("ecosystem = ?", c.EcosystemID).Count(&n).Error
	return
}

func (c *Contract) ToMap() (v map[string]string) {
	v = make(map[string]string)
	v["id"] = converter.Int64ToStr(c.ID)
	v["name"] = c.Name
	v["value"] = c.Value
	v["wallet_id"] = converter.Int64ToStr(c.WalletID)
	v["token_id"] = converter.Int64ToStr(c.TokenID)
	v["conditions"] = c.Conditions
	v["app_id"] = converter.Int64ToStr(c.AppID)
	v["ecosystem_id"] = converter.Int64ToStr(c.EcosystemID)
	return
}

// GetByApp returns all contracts belonging to selected app
func (c *Contract) GetByApp(appID int64) ([]Contract, error) {
	var result []Contract
	err := DBConn.Select("id, name").Where("app_id = ?", appID).Find(&result).Error
	return result, err
}
