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

package custom

import (
	"errors"

	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/AplaProject/go-apla/packages/utils/tx"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

const (
	firstEcosystemID = 1
	firstAppID       = 1
)

// FirstBlockParser is parser wrapper
type FirstBlockTransaction struct {
	Logger        *log.Entry
	DbTransaction *model.DbTransaction
	Data          interface{}
}

// ErrFirstBlockHostIsEmpty host for first block is not specified
var ErrFirstBlockHostIsEmpty = errors.New("FirstBlockHost is empty")

// Init first block
func (t *FirstBlockTransaction) Init() error {
	return nil
}

// Validate first block
func (t *FirstBlockTransaction) Validate() error {
	return nil
}

// Action is fires first block
func (t *FirstBlockTransaction) Action() error {
	logger := t.Logger
	data := t.Data.(*consts.FirstBlock)
	keyID := crypto.Address(data.PublicKey)
	err := model.ExecSchemaEcosystem(nil, firstEcosystemID, keyID, ``, keyID, firstAppID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("executing ecosystem schema")
		return utils.ErrInfo(err)
	}

	amount := decimal.New(consts.FounderAmount, int32(consts.MoneyDigits)).String()

	commission := &model.SystemParameter{Name: `commission_wallet`}
	if err = commission.SaveArray([][]string{{"1", converter.Int64ToStr(keyID)}}); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("saving commission_wallet array")
		return utils.ErrInfo(err)
	}

	err = model.GetDB(t.DbTransaction).Exec(`update "1_system_parameters" SET value = ? where name = 'test'`, data.Test).Error
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating test parameter")
		return utils.ErrInfo(err)
	}

	err = model.GetDB(t.DbTransaction).Exec(`Update "1_system_parameters" SET value = ? where name = 'private_blockchain'`, data.PrivateBlockchain).Error
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating private_blockchain")
		return utils.ErrInfo(err)
	}

	if err = syspar.SysUpdate(t.DbTransaction); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating syspar")
		return utils.ErrInfo(err)
	}

	err = model.GetDB(t.DbTransaction).Exec(`insert into "1_keys" (id,account,pub,amount) values(?,?,?,?)`,
		keyID, converter.AddressToString(keyID), data.PublicKey, amount).Error
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("inserting key")
		return utils.ErrInfo(err)
	}
	id, err := model.GetNextID(t.DbTransaction, "1_pages")
	if err != nil {
		return utils.ErrInfo(err)
	}
	err = model.GetDB(t.DbTransaction).Exec(`insert into "1_pages" (id,name,menu,value,conditions) values(?, 'default_page',
		  'default_menu', ?, 'ContractConditions("@1DeveloperCondition")')`,
		id, syspar.SysString(`default_ecosystem_page`)).Error
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("inserting default page")
		return utils.ErrInfo(err)
	}
	id, err = model.GetNextID(t.DbTransaction, "1_menu")
	if err != nil {
		return utils.ErrInfo(err)
	}
	err = model.GetDB(t.DbTransaction).Exec(`insert into "1_menu" (id,name,value,title,conditions) values(?, 'default_menu', ?, ?, 'ContractAccess("@1EditMenu")')`,
		id, syspar.SysString(`default_ecosystem_menu`), `default`).Error
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("inserting default menu")
		return utils.ErrInfo(err)
	}
	err = smart.LoadContract(t.DbTransaction, 1)
	if err != nil {
		return utils.ErrInfo(err)
	}
	syspar.SetFirstBlockData(data)
	return nil
}

// Rollback first block
func (t *FirstBlockTransaction) Rollback() error {
	return nil
}

// Header is returns first block header
func (t FirstBlockTransaction) Header() *tx.Header {
	return nil
}
