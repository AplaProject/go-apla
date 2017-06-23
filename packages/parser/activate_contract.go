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

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/shopspring/decimal"
)

// ActivateContractInit initialize ActivateContract transaction
func (p *Parser) ActivateContractInit() error {

	fields := []map[string]string{{"global": "int64"}, {"id": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

// ActivateContractFront checks conditions of ActivateContract transaction
func (p *Parser) ActivateContractFront() error {
	err := p.generalCheck(`activate_contract`)
	if err != nil {
		return p.ErrInfo(err)
	}

	forSign := fmt.Sprintf("%s,%s,%d,%d,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxCitizenID,
		p.TxStateID, p.TxMap["global"], p.TxMap["id"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	prefix := `global`
	if p.TxMaps.Int64["global"] == 0 {
		prefix = p.TxStateIDStr
	}
	if len(p.TxMaps.String["id"]) == 0 {
		return p.ErrInfo("incorrect contract id")
	}
	if p.TxMaps.String["id"][0] > '9' {
		p.TxMaps.String["id"], err = p.Single(`SELECT id FROM "`+prefix+`_smart_contracts" WHERE name = ?`, p.TxMaps.String["id"]).String()
		if len(p.TxMaps.String["id"]) == 0 {
			return p.ErrInfo("incorrect contract name")
		}
	}
	active, err := p.Single(`SELECT active FROM "`+prefix+`_smart_contracts" WHERE id = ?`, p.TxMaps.String["id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	if active == `1` {
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
	p.TxMaps.String["activate_cost"] = cost.String()
	return nil
}

// ActivateContract is a main function of ActivateContract transaction
func (p *Parser) ActivateContract() error {
	prefix := `global`
	if p.TxMaps.Int64["global"] == 0 {
		prefix = p.TxStateIDStr
	}
	wallet := p.TxWalletID
	if wallet == 0 {
		wallet = p.TxCitizenID
	}
	egs := p.TxMaps.String["activate_cost"]
	if _, err := p.selectiveLoggingAndUpd([]string{`-amount`}, []interface{}{egs}, `dlt_wallets`, []string{`wallet_id`},
		[]string{converter.Int64ToStr(wallet)}, true); err != nil {
		return err
	}
	if _, err := p.selectiveLoggingAndUpd([]string{`+amount`}, []interface{}{egs}, `dlt_wallets`, []string{`wallet_id`},
		[]string{converter.Int64ToStr(p.BlockData.WalletId)}, true); err != nil {
		return err
	}
	if _, err := p.selectiveLoggingAndUpd([]string{`active`}, []interface{}{1}, prefix+`_smart_contracts`, []string{`id`},
		[]string{p.TxMaps.String["id"]}, true); err != nil {
		return err
	}
	smart.ActivateContract(converter.StrToInt64(p.TxMaps.String["id"]), prefix, true)
	return nil
}

// ActivateContractRollback rollbacks ActivateContract transaction
func (p *Parser) ActivateContractRollback() error {
	return p.autoRollback()
}
