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
	"strconv"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type EditContractParser struct {
	*Parser
	EditContract *tx.EditContract
}

func (p *EditContractParser) Init() error {
	editContract := &tx.EditContract{}
	if err := msgpack.Unmarshal(p.TxBinaryData, editContract); err != nil {
		return p.ErrInfo(err)
	}
	p.EditContract = editContract
	return nil
}

func (p *EditContractParser) Validate() error {
	err := p.generalCheck(`edit_contract`, &p.EditContract.Header, map[string]string{"conditions": p.EditContract.Conditions})
	if err != nil {
		return p.ErrInfo(err)
	}

	// must be supplemented
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.EditContract.ForSign(), p.EditContract.BinSignatures, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	prefix, err := GetTablePrefix(p.EditContract.Global, p.EditContract.Header.StateID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(p.EditContract.Conditions) > 0 {
		if err := smart.CompileEval(string(p.EditContract.Conditions), uint32(p.EditContract.Header.StateID)); err != nil {
			return p.ErrInfo(err)
		}
	}
	sc := &model.SmartContract{}
	sc.SetTablePrefix(prefix)
	contractID, err := strconv.ParseInt(p.EditContract.Id, 10, 64)
	if err != nil {
		logger.LogInfo(consts.StrtoInt64Error, p.EditContract.Id)
	}
	err = sc.GetByID(contractID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(sc.Conditions) > 0 {
		ret, err := p.EvalIf(sc.Conditions)
		if err != nil {
			return err
		}
		if !ret {
			if err = p.AccessRights(`changing_smart_contracts`, false); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *EditContractParser) Action() error {
	prefix, err := GetTablePrefix(p.EditContract.Global, p.EditContract.Header.StateID)
	if err != nil {
		return p.ErrInfo(err)
	}
	sc := &model.SmartContract{}
	sc.SetTablePrefix(prefix)
	contractID, err := strconv.ParseInt(p.EditContract.Id, 10, 64)
	if err != nil {
		logger.LogInfo(consts.StrtoInt64Error, p.EditContract.Id)
	}
	err = sc.GetByID(contractID)
	if err != nil {
		return p.ErrInfo(err)
	}
	active := sc.Active == `1`
	root, err := smart.CompileBlock(p.EditContract.Value, prefix, false, contractID)
	if err != nil {
		return p.ErrInfo(err)
	}
	_, _, err = p.selectiveLoggingAndUpd([]string{"value", "conditions"}, []interface{}{p.EditContract.Value, p.EditContract.Conditions}, prefix+"_smart_contracts", []string{"id"}, []string{p.EditContract.Id}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	for i, item := range root.Children {
		if item.Type == script.ObjContract {
			root.Children[i].Info.(*script.ContractInfo).TableID = sc.ID
			root.Children[i].Info.(*script.ContractInfo).Active = active
		}
	}
	smart.FlushBlock(root)
	return nil
}

func (p *EditContractParser) Rollback() error {
	return p.autoRollback()
}

func (p EditContractParser) Header() *tx.Header {
	return &p.EditContract.Header
}
