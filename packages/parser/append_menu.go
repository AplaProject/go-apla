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
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type AppendMenuParser struct {
	*Parser
	AppendMenu *tx.AppendMenu
}

func (p *AppendMenuParser) Init() error {
	appendMenu := &tx.AppendMenu{}
	if err := msgpack.Unmarshal(p.TxBinaryData, appendMenu); err != nil {
		return p.ErrInfo(err)
	}
	p.AppendMenu = appendMenu
	return nil
}

func (p *AppendMenuParser) Validate() error {
	err := p.generalCheck(`edit_menu`, &p.AppendMenu.Header, map[string]string{})
	if err != nil {
		return p.ErrInfo(err)
	}

	// must be supplemented
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.AppendMenu.ForSign(), p.AppendMenu.Header.BinSignatures, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if err = p.AccessChange(`menu`, p.AppendMenu.Name, p.AppendMenu.Global, p.AppendMenu.StateID); err != nil {
		if p.AccessRights(`changing_menu`, false) != nil {
			return err
		}
	}
	return nil
}

func (p *AppendMenuParser) Action() error {
	prefix, err := GetTablePrefix(p.AppendMenu.Global, p.AppendMenu.Header.StateID)
	if err != nil {
		return p.ErrInfo(err)
	}
	log.Debug("value page", p.AppendMenu.Value)
	menu := &model.Menu{}
	menu.SetTablePrefix(prefix)
	err = menu.Get(p.AppendMenu.Name)
	if err != nil {
		return p.ErrInfo(err)
	}
	page := menu.Value
	new := page + "\r\n" + p.AppendMenu.Value
	_, _, err = p.selectiveLoggingAndUpd([]string{"value"}, []interface{}{new}, prefix+"_menu", []string{"name"}, []string{p.AppendMenu.Name}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *AppendMenuParser) Rollback() error {
	return p.autoRollback()
}

func (p AppendMenuParser) Header() *tx.Header {
	return &p.AppendMenu.Header
}
