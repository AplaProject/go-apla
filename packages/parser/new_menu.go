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
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type NewMenuParser struct {
	*Parser
	NewMenu *tx.NewMenu
}

func (p *NewMenuParser) Init() error {
	newMenu := &tx.NewMenu{}
	if err := msgpack.Unmarshal(p.TxBinaryData, newMenu); err != nil {
		return p.ErrInfo(err)
	}
	p.NewMenu = newMenu
	return nil

}

func (p *NewMenuParser) Validate() error {
	err := p.generalCheck(`new_menu`, &p.NewMenu.Header, map[string]string{"conditions": p.NewMenu.Conditions})
	if err != nil {
		return p.ErrInfo(err)
	}

	// must be supplemented
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.NewMenu.ForSign(), p.NewMenu.BinSignatures, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *NewMenuParser) Action() error {
	prefix, err := GetTablePrefix(p.NewMenu.Global, p.NewMenu.Header.StateID)
	if err != nil {
		return p.ErrInfo(err)
	}
	_, err = p.selectiveLoggingAndUpd([]string{"name", "value", "conditions"}, []interface{}{p.NewMenu.Name, p.NewMenu.Value, p.NewMenu.Conditions}, prefix+"_menu", nil, nil, true)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *NewMenuParser) Rollback() error {
	return p.autoRollback()
}

func (p *NewMenuParser) Header() *tx.Header {
	return &p.NewMenu.Header
}
