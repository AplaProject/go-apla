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
	"encoding/json"
	"fmt"

	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type NewLangParser struct {
	*Parser
	NewLang *tx.EditNewLang
}

func (p *NewLangParser) Init() error {
	newLang := &tx.EditNewLang{}
	if err := msgpack.Unmarshal(p.BinaryData, newLang); err != nil {
		return p.ErrInfo(err)
	}
	p.NewLang = newLang
	return nil
}

func (p *NewLangParser) Validate() error {
	err := p.generalCheck(`new_lang`, &p.NewLang.Header)
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
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.NewLang.ForSign(), p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if err = p.AccessRights(`changing_language`, false); err != nil {
		return p.ErrInfo(err)
	}
	prefix := utils.Int64ToStr(p.NewLang.Header.StateID)
	if len(p.NewLang.Name) == 0 {
		var list map[string]string
		err := json.Unmarshal([]byte(p.NewLang.Trans), &list)
		if err != nil {
			return p.ErrInfo(err)
		}
		if len(list) == 0 {
			return fmt.Errorf(`empty lanuguage resource`)
		}
	} else {
		if exist, err := p.Single(`select name from "`+prefix+"_languages"+`" where name=?`, p.NewLang.Name).String(); err != nil {
			return p.ErrInfo(err)
		} else if len(exist) > 0 {
			return p.ErrInfo(fmt.Sprintf("The language resource %s already exists", p.NewLang.Name))
		}
	}
	return nil
}

func (p *NewLangParser) Action() error {
	prefix := utils.Int64ToStr(p.NewLang.Header.StateID)
	if len(p.NewLang.Name) == 0 {
		var list map[string]string
		json.Unmarshal([]byte(p.NewLang.Trans), &list)
		for name, res := range list {
			if exist, err := p.Single(`select name from "`+prefix+"_languages"+`" where name=?`, name).String(); err != nil {
				return p.ErrInfo(err)
			} else if len(exist) == 0 {
				_, err := p.selectiveLoggingAndUpd([]string{"name", "res"}, []interface{}{name, res}, prefix+"_languages", nil, nil, true)
				if err != nil {
					return p.ErrInfo(err)
				}
				utils.UpdateLang(int(p.NewLang.Header.StateID), name, res)
			}
		}
	} else {
		_, err := p.selectiveLoggingAndUpd([]string{"name", "res"}, []interface{}{p.NewLang.Name,
			p.NewLang.Trans}, prefix+"_languages", nil, nil, true)
		if err != nil {
			return p.ErrInfo(err)
		}
		utils.UpdateLang(int(p.NewLang.Header.StateID), p.NewLang.Name, p.NewLang.Trans)
	}
	return nil
}

func (p *NewLangParser) Rollback() error {
	return p.autoRollback()
}

func (p NewLangParser) Header() *tx.Header {
	return &p.NewLang.Header
}
