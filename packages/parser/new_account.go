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
	"encoding/hex"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type NewAccountParser struct {
	*Parser
	NewAccount *tx.NewAccount
}

func (p *NewAccountParser) Init() error {
	newAccount := &tx.NewAccount{}
	if err := msgpack.Unmarshal(p.TxBinaryData, newAccount); err != nil {
		return p.ErrInfo(err)
	}
	p.NewAccount = newAccount
	return nil
}

func (p *NewAccountParser) Validate() error {
	p.PublicKeys = append(p.PublicKeys, p.NewAccount.PublicKey)
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.NewAccount.ForSign(), p.NewAccount.BinSignatures, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	return nil
}

func (p *NewAccountParser) Action() error {
	_, err := p.selectiveLoggingAndUpd([]string{"public_key_0"}, []interface{}{hex.EncodeToString(p.NewAccount.PublicKey)},
		"dlt_wallets", []string{"wallet_id"}, []string{converter.Int64ToStr(p.TxCitizenID)}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	_, err = p.selectiveLoggingAndUpd([]string{"citizen_id", "amount"}, []interface{}{p.TxCitizenID, 0},
		converter.UInt32ToStr(p.TxStateID)+"_accounts", nil, nil, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *NewAccountParser) Rollback() error {
	return p.autoRollback()
}

func (p NewAccountParser) Header() *tx.Header {
	return &p.NewAccount.Header
}
