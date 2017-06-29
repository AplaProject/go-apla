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

	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type EditWalletParser struct {
	*Parser
	EditWallet *tx.EditWallet
}

func (p *EditWalletParser) Init() error {
	editWallet := &tx.EditWallet{}
	if err := msgpack.Unmarshal(p.TxBinaryData, editWallet); err != nil {
		return p.ErrInfo(err)
	}
	p.EditWallet = editWallet
	return nil
}

func (p *EditWalletParser) checkContract(name string) string {
	name = script.StateName(uint32(p.EditWallet.Header.StateID), name)
	if smart.GetContract(name, 0) == nil {
		return ``
	}
	return name
}

func (p *EditWalletParser) Validate() error {
	err := p.generalCheck(`edit_wallet`, &p.EditWallet.Header, map[string]string{"conditions": p.EditWallet.Conditions})
	if err != nil {
		return p.ErrInfo(err)
	}

	wallet := p.TxWalletID
	if wallet == 0 {
		wallet = p.TxCitizenID
	}
	// must be supplemented
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.EditWallet.ForSign(), p.EditWallet.BinSignatures, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if len(p.EditWallet.Conditions) > 0 {
		if err := smart.CompileEval(string(p.EditWallet.Conditions), uint32(p.EditWallet.Header.StateID)); err != nil {
			return p.ErrInfo(err)
		}
	}
	id := utils.StrToInt64(string(p.EditWallet.WalletID))
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
	if len(p.EditWallet.SpendingContract) > 0 {
		if len(p.checkContract(string(p.EditWallet.SpendingContract))) == 0 {
			return fmt.Errorf(`Cannot find %s contract`, string(p.EditWallet.SpendingContract))
		}
	}
	return nil
}

func (p *EditWalletParser) Action() error {
	var contract string

	if len(p.EditWallet.SpendingContract) > 0 {
		contract = p.checkContract(string(p.EditWallet.SpendingContract))
	}
	_, err := p.selectiveLoggingAndUpd([]string{"spending_contract", "conditions_change"},
		[]interface{}{contract, string(p.EditWallet.Conditions)}, "dlt_wallets",
		[]string{"wallet_id"}, []string{string(p.EditWallet.WalletID)}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *EditWalletParser) Rollback() error {
	return p.autoRollback()
}

func (p *EditWalletParser) Header() *tx.Header {
	return &p.EditWallet.Header
}
