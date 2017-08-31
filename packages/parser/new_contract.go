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
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type NewContractParser struct {
	*Parser
	NewContract    *tx.NewContract
	walletContract int64
}

func (p *NewContractParser) Init() error {
	newContract := &tx.NewContract{}
	if err := msgpack.Unmarshal(p.TxBinaryData, newContract); err != nil {
		return p.ErrInfo(err)
	}
	p.NewContract = newContract
	return nil
}

func (p *NewContractParser) Validate() error {
	err := p.generalCheck(`new_contract`, &p.NewContract.Header, map[string]string{"conditions": p.NewContract.Conditions})
	if err != nil {
		return p.ErrInfo(err)
	}

	// Check the system limits. You can not send more than X time a day this TX
	// ...

	// Check InputData
	if len(p.NewContract.Wallet) > 0 {
		p.walletContract = converter.StringToAddress(p.NewContract.Wallet)
		if p.walletContract == 0 {
			return p.ErrInfo(fmt.Errorf(`wrong wallet %s`, p.NewContract.Wallet))
		}
	}
	verifyData := map[string][]interface{}{"int64": []interface{}{p.NewContract.Global}, "string": []interface{}{p.NewContract.Name}}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// must be supplemented
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.NewContract.ForSign(), p.NewContract.BinSignatures, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	prefix, err := GetTablePrefix(p.NewContract.Global, p.NewContract.Header.StateID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(p.NewContract.Conditions) > 0 {
		if err := smart.CompileEval(string(p.NewContract.Conditions), uint32(p.NewContract.UserID)); err != nil {
			return p.ErrInfo(err)
		}
	}

	sc := &model.SmartContract{}
	sc.SetTablePrefix(prefix)
	if exist, err := sc.ExistsByName(p.NewContract.Name); err != nil {
		return p.ErrInfo(err)
	} else if exist {
		return p.ErrInfo(fmt.Sprintf("The contract %s already exists", p.NewContract.Name))
	}
	return nil
}

func (p *NewContractParser) Action() error {
	prefix, err := GetTablePrefix(p.NewContract.Global, p.NewContract.Header.StateID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if p.walletContract == 0 {
		p.walletContract = p.NewContract.UserID
	}
	root, err := smart.CompileBlock(p.NewContract.Value, prefix, false, 0)
	if err != nil {
		return p.ErrInfo(err)
	}

	_, tblid, err := p.selectiveLoggingAndUpd([]string{"name", "value", "conditions", "wallet_id"},
		[]interface{}{p.NewContract.Name, p.NewContract.Value, p.NewContract.Conditions,
			p.walletContract}, prefix+"_smart_contracts", nil, nil, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	for i, item := range root.Children {
		if item.Type == script.ObjContract {
			tableID, err := strconv.ParseInt(tblid, 10, 64)
			if err != nil {
				logger.LogInfo(consts.StrToIntError, tblid)
			}
			root.Children[i].Info.(*script.ContractInfo).TableID = tableID
		}
	}

	smart.FlushBlock(root)
	return nil
}

func (p *NewContractParser) Rollback() error {
	return p.autoRollback()
}

func (p NewContractParser) Header() *tx.Header {
	return &p.NewContract.Header
}
