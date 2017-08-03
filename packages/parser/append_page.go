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
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type AppendPageParser struct {
	*Parser
	AppendPage *tx.AppendPage
}

func (p *AppendPageParser) Init() error {
	appendPage := &tx.AppendPage{}
	if err := msgpack.Unmarshal(p.TxBinaryData, appendPage); err != nil {
		return p.ErrInfo(err)
	}
	p.AppendPage = appendPage
	return nil
}

func (p *AppendPageParser) Validate() error {
	err := p.generalCheck(`edit_page`, &p.AppendPage.Header, map[string]string{})
	if err != nil {
		return p.ErrInfo(err)
	}

	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.AppendPage.ForSign(), p.AppendPage.BinSignatures, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if err = p.AccessChange(`pages`, p.AppendPage.Name, p.AppendPage.Global, p.AppendPage.StateID); err != nil {
		if p.AccessRights(`changing_page`, false) != nil {
			return err
		}
	}
	return nil
}

func (p *AppendPageParser) Action() error {
	prefix, err := GetTablePrefix(p.AppendPage.Global, p.AppendPage.Header.StateID)
	if err != nil {
		return p.ErrInfo(err)
	}
	log.Debug("value page", p.AppendPage.Value)
	page := &model.Page{}
	page.SetTablePrefix(prefix)
	err = page.Get(p.AppendPage.Name)
	if err != nil {
		return p.ErrInfo(err)
	}
	new := strings.Replace(page.Value, "PageEnd:", p.AppendPage.Value, -1) + "\r\nPageEnd:"
	_, _, err = p.selectiveLoggingAndUpd([]string{"value"}, []interface{}{new}, prefix+"_pages", []string{"name"}, []string{p.AppendPage.Name}, true)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *AppendPageParser) Rollback() error {
	return p.autoRollback()
}

func (p AppendPageParser) Header() *tx.Header {
	return &p.AppendPage.Header
}
