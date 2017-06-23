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
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// EditWalletInit initializes EditWallet transaction
func (p *Parser) EditWalletInit() error {

	fields := []map[string]string{{"id": "int64"}, {"spending_contract": "string"},
		{"conditions_change": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) checkContract(name string) string {
	name = script.StateName(p.TxStateID, name)
	if smart.GetContract(name, 0) == nil {
		return ``
	}
	return name
}

// EditWalletFront checks conditions of EditWallet transaction
func (p *Parser) EditWalletFront() error {

	err := p.generalCheck(`edit_wallet`)
	if err != nil {
		return p.ErrInfo(err)
	}

	// Check the system limits. You can not send more than X time a day this TX
	// ...

	// Check InputData
	verifyData := map[string]string{}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	wallet := p.TxWalletID
	if wallet == 0 {
		wallet = p.TxCitizenID
	}
	// must be supplemented
	forSign := fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], wallet,
		p.TxStateID, p.TxMap["id"], p.TxMap["spending_contract"], p.TxMap["conditions_change"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if len(p.TxMap["conditions_change"]) > 0 {
		if err := smart.CompileEval(string(p.TxMap["conditions_change"]), uint32(p.TxStateID)); err != nil {
			return p.ErrInfo(err)
		}
	}
	id := converter.StrToInt64(string(p.TxMap["id"]))
	conditions, err := p.Single(`SELECT conditions_change FROM "dlt_wallets" WHERE wallet_id = ?`, id).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(conditions) > 0 && conditions != `NULL` {
		ret, err := p.EvalIf(conditions)
		if err != nil {
			return p.ErrInfo(err)
		}
		if !ret {
			return fmt.Errorf(`Access denied`)
		}
	} else if id != wallet {
		return fmt.Errorf(`Access denied`)
	}
	if len(p.TxMap["spending_contract"]) > 0 {
		if len(p.checkContract(string(p.TxMap["spending_contract"]))) == 0 {
			return fmt.Errorf(`Cannot find %s contract`, string(p.TxMap["spending_contract"]))
		}
	}
	return nil
}

// EditWallet proceeds EditWallet transaction
func (p *Parser) EditWallet() error {
	var contract string

	if len(p.TxMap["spending_contract"]) > 0 {
		contract = p.checkContract(string(p.TxMap["spending_contract"]))
	}
	_, err := p.selectiveLoggingAndUpd([]string{"spending_contract", "conditions_change"},
		[]interface{}{contract, string(p.TxMap["conditions_change"])}, "dlt_wallets",
		[]string{"wallet_id"}, []string{string(p.TxMap["id"])}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

// EditWalletRollback rollbacks EditWallet transaction
func (p *Parser) EditWalletRollback() error {
	return p.autoRollback()
}
