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

type EditLangParser struct {
	*Parser
	EditLang *tx.EditNewLang
}

func (p *EditLangParser) Init() error {
	editLang := &tx.EditNewLang{}
	if err := msgpack.Unmarshal(p.BinaryData, editLang); err != nil {
		return p.ErrInfo(err)
	}
	p.EditLang = editLang
	return nil
}

func (p *EditLangParser) Validate() error {
	err := p.generalCheck(`edit_lang`, &p.EditLang.Header)
	if err != nil {
		return p.ErrInfo(err)
	}
	// Check InputData
	verifyData := map[string]string{}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// must be supplemented
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.EditLang.ForSign(), p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if err = p.AccessRights(`changing_language`, false); err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *EditLangParser) Action() error {
	prefix := utils.Int64ToStr(p.EditLang.Header.StateID)
	_, err := p.selectiveLoggingAndUpd([]string{"res"}, []interface{}{p.EditLang.Trans}, prefix+"_languages", []string{"name"}, []string{p.EditLang.Name}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	utils.UpdateLang(int(p.EditLang.Header.StateID), p.EditLang.Name, p.EditLang.Trans)
	return nil
}

func (p *EditLangParser) Rollback() error {
	return p.autoRollback()
}

func (p EditLangParser) Header() *tx.Header {
	return &p.EditLang.Header
}
