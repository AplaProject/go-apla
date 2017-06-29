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

type EditMenuParser struct {
	*Parser
	EditMenu *tx.EditMenu
}

func (p *EditMenuParser) Init() error {
	editMenu := &tx.EditMenu{}
	if err := msgpack.Unmarshal(p.TxBinaryData, editMenu); err != nil {
		return p.ErrInfo(err)
	}
	p.EditMenu = editMenu
	return nil
}

func (p *EditMenuParser) Validate() error {
	err := p.generalCheck(`edit_menu`, &p.EditMenu.Header, map[string]string{"conditions": p.EditMenu.Conditions})
	if err != nil {
		return p.ErrInfo(err)
	}

	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.EditMenu.ForSign(), p.EditMenu.BinSignatures, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if len(p.EditMenu.Conditions) > 0 {
		if err := smart.CompileEval(string(p.EditMenu.Conditions), uint32(p.EditMenu.Header.StateID)); err != nil {
			return p.ErrInfo(err)
		}
	}

	if err = p.AccessChange(`menu`, p.EditMenu.Name, p.EditMenu.Global, p.EditMenu.StateID); err != nil {
		if p.AccessRights(`changing_menu`, false) != nil {
			return err
		}
	}

	return nil
}

func (p *EditMenuParser) Action() error {
	prefix, err := GetTablePrefix(p.EditMenu.Global, p.EditMenu.Header.StateID)
	if err != nil {
		return p.ErrInfo(err)
	}
	_, err = p.selectiveLoggingAndUpd([]string{"value", "conditions"}, []interface{}{p.EditMenu.Value, p.EditMenu.Conditions}, prefix+"_menu", []string{"name"}, []string{p.EditMenu.Name}, true)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *EditMenuParser) Rollback() error {
	return p.autoRollback()
}

func (p EditMenuParser) Header() *tx.Header {
	return &p.EditMenu.Header
}
