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
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type EditSignParser struct {
	*Parser
	EditSign *tx.EditNewSign
}

func (p *EditSignParser) Init() error {
	editSign := &tx.EditNewSign{}
	if err := msgpack.Unmarshal(p.TxBinaryData, editSign); err != nil {
		return p.ErrInfo(err)
	}
	p.EditSign = editSign
	return nil
}

func (p *EditSignParser) Validate() error {
	err := p.generalCheck(`edit_sign`, &p.EditSign.Header, map[string]string{"conditions": p.EditSign.Conditions})
	if err != nil {
		return p.ErrInfo(err)
	}

	// must be supplemented
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.EditSign.ForSign(), p.EditSign.BinSignatures, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	prefix := utils.Int64ToStr(int64(p.EditSign.Header.StateID))
	if len(p.EditSign.Conditions) > 0 {
		if err := smart.CompileEval(string(p.EditSign.Conditions), uint32(p.EditSign.Header.StateID)); err != nil {
			return p.ErrInfo(err)
		}
	}
	conditions, err := p.Single(`SELECT conditions FROM "`+prefix+`_signatures" WHERE name = ?`, p.EditSign.Name).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(conditions) > 0 {
		ret, err := p.EvalIf(conditions)
		if err != nil {
			return p.ErrInfo(err)
		}
		if !ret {
			if err = p.AccessRights(`changing_signatures`, false); err != nil {
				return p.ErrInfo(err)
			}
		}
	}

	return nil
}

func (p *EditSignParser) Action() error {
	prefix, err := GetTablePrefix(p.EditSign.Global, p.EditSign.Header.StateID)
	if err != nil {
		return p.ErrInfo(err)
	}
	_, err = p.selectiveLoggingAndUpd([]string{"value", "conditions"}, []interface{}{p.EditSign.Value,
		p.EditSign.Conditions}, prefix+"_signatures", []string{"name"}, []string{p.EditSign.Name}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *EditSignParser) Rollback() error {
	return p.autoRollback()
}

func (p EditSignParser) Header() *tx.Header {
	return &p.EditSign.Header
}
