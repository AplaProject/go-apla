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
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type NewPageParser struct {
	*Parser
	NewPage *tx.NewPage
}

func (p *NewPageParser) Init() error {
	newPage := &tx.NewPage{}
	if err := msgpack.Unmarshal(p.TxBinaryData, newPage); err != nil {
		return p.ErrInfo(err)
	}
	p.NewPage = newPage
	return nil
}

func (p *NewPageParser) Validate() error {
	err := p.generalCheck(`new_page`, &p.NewPage.Header)
	if err != nil {
		return p.ErrInfo(err)
	}
	if strings.HasPrefix(string(p.NewPage.Name), `sys-`) || strings.HasPrefix(string(p.NewPage.Name), `app-`) {
		return fmt.Errorf(`The name cannot start with sys- or app-`)
	}
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.NewPage.ForSign(), p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	return nil
}

func (p *NewPageParser) Action() error {
	prefix, err := GetTablePrefix(p.NewPage.Global, p.NewPage.Header.StateID)
	if err != nil {
		return p.ErrInfo(err)
	}
	_, err = p.selectiveLoggingAndUpd([]string{"name", "value", "menu", "conditions"}, []interface{}{p.NewPage.Name, p.NewPage.Value, p.NewPage.Menu, p.NewPage.Conditions}, prefix+"_pages", nil, nil, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *NewPageParser) Rollback() error {
	return p.autoRollback()
}

func (p NewPageParser) Header() *tx.Header {
	return &p.NewPage.Header
}
