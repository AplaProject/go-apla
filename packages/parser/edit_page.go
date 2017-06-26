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

type EditPageParser struct {
	*Parser
	EditPage *tx.EditPage
}

func (p *EditPageParser) Init() error {
	editPage := &tx.EditPage{}
	if err := msgpack.Unmarshal(p.TxBinaryData, editPage); err != nil {
		return p.ErrInfo(err)
	}
	p.EditPage = editPage
	return nil
}

func (p *EditPageParser) Validate() error {
	p.TxMap["conditions"] = []byte(p.EditPage.Conditions)
	err := p.generalCheck(`edit_page`, &p.EditPage.Header)
	if err != nil {
		return p.ErrInfo(err)
	}

	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.EditPage.ForSign(), p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if len(p.EditPage.Conditions) > 0 {
		if err := smart.CompileEval(string(p.EditPage.Conditions), uint32(p.EditPage.Header.StateID)); err != nil {
			return p.ErrInfo(err)
		}
	}
	if err = p.AccessChange(`pages`, p.EditPage.Name); err != nil {
		if p.AccessRights(`changing_page`, false) != nil {
			return err
		}
	}
	return nil
}

func (p *EditPageParser) Action() error {
	prefix, err := GetTablePrefix(p.EditPage.Global, p.EditPage.Header.StateID)
	if err != nil {
		return p.ErrInfo(err)
	}

	log.Debug("value page", p.EditPage.Value)
	_, err = p.selectiveLoggingAndUpd([]string{"value", "menu", "conditions"}, []interface{}{p.EditPage.Value, p.EditPage.Menu, p.EditPage.Conditions}, prefix+"_pages", []string{"name"}, []string{p.EditPage.Name}, true)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *EditPageParser) Rollback() error {
	return p.autoRollback()
}

func (p *EditPageParser) Header() *tx.Header {
	return &p.EditPage.Header
}
