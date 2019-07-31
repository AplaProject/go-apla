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
