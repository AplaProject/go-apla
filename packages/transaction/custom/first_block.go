// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package custom

import (
	"errors"

	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/smart"
	"github.com/GenesisKernel/go-genesis/packages/utils"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

const firstEcosystemID = 1

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
	err := model.ExecSchemaEcosystem(nil, firstEcosystemID, keyID, ``, keyID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("executing ecosystem schema")
		return utils.ErrInfo(err)
	}

	sp := &model.StateParameter{}
	sp.SetTablePrefix(converter.IntToStr(firstEcosystemID))
	_, err = sp.Get(nil, model.ParamMoneyDigit)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting ecosystem param")
		return err
	}
	amount := decimal.New(consts.FounderAmount, int32(converter.StrToInt64(sp.Value))).String()

	err = model.GetDB(t.DbTransaction).Exec(`insert into "1_keys" (id,pub,amount) values(?, ?,?)`,
		keyID, data.PublicKey, amount).Error
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("inserting default page")
		return utils.ErrInfo(err)
	}
	err = model.GetDB(t.DbTransaction).Exec(`insert into "1_pages" (id,name,menu,value,conditions) values('1', 'default_page',
		  'default_menu', ?, 'ContractAccess("@1EditPage")')`, syspar.SysString(`default_ecosystem_page`)).Error
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("inserting default page")
		return utils.ErrInfo(err)
	}
	err = model.GetDB(t.DbTransaction).Exec(`insert into "1_menu" (id,name,value,title,conditions) values('1', 'default_menu', ?, ?, 'ContractAccess("@1EditMenu")')`,
		syspar.SysString(`default_ecosystem_menu`), `default`).Error
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("inserting default menu")
		return utils.ErrInfo(err)
	}
	err = smart.LoadContract(t.DbTransaction, `1`)
	if err != nil {
		return utils.ErrInfo(err)
	}
	commission := &model.SystemParameter{Name: `commission_wallet`}
	if err = commission.SaveArray([][]string{{"1", converter.Int64ToStr(keyID)}}); err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("saving commission_wallet array")
		return utils.ErrInfo(err)
	}
	if err = syspar.SysUpdate(nil); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating syspar")
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
