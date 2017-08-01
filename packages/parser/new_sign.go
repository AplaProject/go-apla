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

	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type NewSignParser struct {
	*Parser
	NewSign *tx.EditNewSign
}

func (p *NewSignParser) Init() error {
	newSign := &tx.EditNewSign{}
	if err := msgpack.Unmarshal(p.TxBinaryData, newSign); err != nil {
		return p.ErrInfo(err)
	}
	p.NewSign = newSign
	return nil
}

func (p *NewSignParser) Validate() error {
	err := p.generalCheck(`new_sign`, &p.NewSign.Header, map[string]string{})
	if err != nil {
		return p.ErrInfo(err)
	}

	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.NewSign.ForSign(), p.NewSign.BinSignatures, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if err = p.AccessRights(`changing_signature`, false); err != nil {
		return p.ErrInfo(err)
	}
	prefix, err := GetTablePrefix(p.NewSign.Global, p.NewSign.Header.StateID)
	if err != nil {
		return p.ErrInfo(err)
	}
	sign := &model.Signatures{}
	sign.SetTableName(prefix + "_signatures")
	if exist, err := sign.ExistsByName(p.NewSign.Name); err != nil {
		return p.ErrInfo(err)
	} else if exist {
		return p.ErrInfo(fmt.Sprintf("The signature %s already exists", p.NewSign.Name))
	}
	return nil
}

func (p *NewSignParser) Action() error {
	prefix, err := GetTablePrefix(p.NewSign.Global, p.NewSign.Header.StateID)
	if err != nil {
		return p.ErrInfo(err)
	}
	_, _, err = p.selectiveLoggingAndUpd([]string{"name", "value", "conditions"}, []interface{}{p.NewSign.Name, p.NewSign.Value, p.NewSign.Conditions}, prefix+"_signatures", nil, nil, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *NewSignParser) Rollback() error {
	return p.autoRollback()
}

func (p NewSignParser) Header() *tx.Header {
	return &p.NewSign.Header
}
