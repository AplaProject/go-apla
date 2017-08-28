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

package parser

import (
	"fmt"
	"strconv"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/shopspring/decimal"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type ActivateContractParser struct {
	*Parser
	ActivateContract *tx.ActivateContract
	activateCost     string
}

func (p *ActivateContractParser) Init() error {
	activateContract := &tx.ActivateContract{}
	if err := msgpack.Unmarshal(p.TxBinaryData, activateContract); err != nil {
		return p.ErrInfo(err)
	}
	p.ActivateContract = activateContract
	return nil
}

func (p *ActivateContractParser) Validate() error {
	err := p.generalCheck(`activate_contract`, &p.ActivateContract.Header, map[string]string{})
	if err != nil {
		return p.ErrInfo(err)
	}

	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.ActivateContract.ForSign(), p.ActivateContract.Header.BinSignatures, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	prefix, err := GetTablePrefix(p.ActivateContract.Global, p.ActivateContract.Header.StateID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(p.ActivateContract.Id) == 0 {
		return p.ErrInfo("incorrect contract id")
	}
	if p.ActivateContract.Id[0] > '9' {
		smartContract := &model.SmartContract{}
		smartContract.SetTablePrefix(prefix)
		err = smartContract.GetByName(p.ActivateContract.Id)
		if smartContract.ID == 0 {
			return p.ErrInfo("incorrect contract name")
		}
		p.ActivateContract.Id = converter.Int64ToStr(smartContract.ID)
	}
	smartContract := &model.SmartContract{}
	smartContract.SetTablePrefix(prefix)
	contractID, err := strconv.ParseInt(p.ActivateContract.Id, 10, 64)
	if err != nil {
		logger.LogInfo(consts.StrtoInt64Error, p.ActivateContract.Id)
	}
	err = smartContract.GetByID(contractID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if smartContract.Active == `1` {
		return p.ErrInfo(fmt.Errorf(`The contract has been already activated`))
	}
	curCost := p.TxUsedCost
	cost, err := p.getEGSPrice(`activate_cost`)
	p.TxUsedCost = curCost
	if err != nil {
		return p.ErrInfo(err)
	}
	if err := p.checkSenderDLT(cost, decimal.New(0, 0)); err != nil {
		return p.ErrInfo(err)
	}
	p.activateCost = cost.String()
	return nil
}

func (p *ActivateContractParser) Action() error {
	prefix, err := GetTablePrefix(p.ActivateContract.Global, p.ActivateContract.Header.StateID)
	if err != nil {
		return p.ErrInfo(err)
	}
	wallet := p.TxWalletID
	if wallet == 0 {
		wallet = p.TxCitizenID
	}
	egs := p.activateCost
	if _, _, err := p.selectiveLoggingAndUpd([]string{`-amount`}, []interface{}{egs}, `dlt_wallets`, []string{`wallet_id`},
		[]string{converter.Int64ToStr(wallet)}, true); err != nil {
		return err
	}
	if _, _, err := p.selectiveLoggingAndUpd([]string{`+amount`}, []interface{}{egs}, `dlt_wallets`, []string{`wallet_id`},
		[]string{converter.Int64ToStr(p.BlockData.WalletID)}, true); err != nil {
		return err
	}
	if _, _, err := p.selectiveLoggingAndUpd([]string{`active`}, []interface{}{1}, prefix+`_smart_contracts`, []string{`id`},
		[]string{p.ActivateContract.Id}, true); err != nil {
		return err
	}
	contractID, err := strconv.ParseInt(p.ActivateContract.Id, 10, 64)
	if err != nil {
		logger.LogInfo(consts.StrtoInt64Error, p.ActivateContract.Id)
	}
	smart.ActivateContract(contractID, prefix, true)
	return nil
}

func (p *ActivateContractParser) Rollback() error {
	return p.autoRollback()
}

func (p *ActivateContractParser) Header() *tx.Header {
	return &p.ActivateContract.Header
}
