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
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/language"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
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
	if err := msgpack.Unmarshal(p.TxBinaryData, newLang); err != nil {
		return p.ErrInfo(err)
	}
	p.NewLang = newLang
	return nil
}

func (p *NewLangParser) Validate() error {
	err := p.generalCheck(`new_lang`, &p.NewLang.Header, map[string]string{})
	if err != nil {
		return p.ErrInfo(err)
	}

	// must be supplemented
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.NewLang.ForSign(), p.NewLang.BinSignatures, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if err = p.AccessRights(`changing_language`, false); err != nil {
		return p.ErrInfo(err)
	}
	prefix := converter.Int64ToStr(p.NewLang.Header.StateID)
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
		lang := &model.Languages{}
		lang.SetTableName(prefix + "_languages")
		if exist, err := lang.IsExistsByName(p.NewLang.Name); err != nil {
			return p.ErrInfo(err)
		} else if exist {
			return p.ErrInfo(fmt.Sprintf("The language resource %s already exists", p.NewLang.Name))
		}
	}
	return nil
}

func (p *NewLangParser) Action() error {
	prefix := converter.Int64ToStr(p.NewLang.Header.StateID)
	if len(p.NewLang.Name) == 0 {
		var list map[string]string
		json.Unmarshal([]byte(p.NewLang.Trans), &list)
		for name, res := range list {
			lang := &model.Languages{}
			lang.SetTableName(prefix + "_languages")
			if exist, err := lang.IsExistsByName(name); err != nil {
				return p.ErrInfo(err)
			} else if !exist {
				_, _, err := p.selectiveLoggingAndUpd([]string{"name", "res"}, []interface{}{name, res}, prefix+"_languages", nil, nil, true)
				if err != nil {
					return p.ErrInfo(err)
				}
				language.UpdateLang(int(p.NewLang.Header.StateID), name, res)
			}
		}
	} else {
		_, _, err := p.selectiveLoggingAndUpd([]string{"name", "res"}, []interface{}{p.NewLang.Name,
			p.NewLang.Trans}, prefix+"_languages", nil, nil, true)
		if err != nil {
			return p.ErrInfo(err)
		}
		language.UpdateLang(int(p.NewLang.Header.StateID), p.NewLang.Name, p.NewLang.Trans)
	}
	return nil
}

func (p *NewLangParser) Rollback() error {
	return p.autoRollback()
}

func (p NewLangParser) Header() *tx.Header {
	return &p.NewLang.Header
}
