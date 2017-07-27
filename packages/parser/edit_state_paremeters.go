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

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	"gopkg.in/vmihailenco/msgpack.v2"
)

/*
Adding state tables should be spelled out in state settings
*/

type EditStateParametersParser struct {
	*Parser
	EditStateParameters *tx.EditStateParameters
}

func (p *EditStateParametersParser) Init() error {
	editStateParameters := &tx.EditStateParameters{}
	if err := msgpack.Unmarshal(p.TxBinaryData, editStateParameters); err != nil {
		return p.ErrInfo(err)
	}
	p.EditStateParameters = editStateParameters
	return nil
}

func (p *EditStateParametersParser) Validate() error {
	err := p.generalCheck(`edit_state_parameters`, &p.EditStateParameters.Header, map[string]string{"conditions": p.EditStateParameters.Conditions})
	if err != nil {
		return p.ErrInfo(err)
	}
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.EditStateParameters.ForSign(), p.EditStateParameters.BinSignatures, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if len(p.EditStateParameters.Conditions) > 0 {
		if err := smart.CompileEval(string(p.EditStateParameters.Conditions), uint32(p.EditStateParameters.Header.StateID)); err != nil {
			return p.ErrInfo(err)
		}
	}
	if p.EditStateParameters.Name == `state_name` {
		if exist, err := IsState(p.EditStateParameters.Value); err != nil {
			return p.ErrInfo(err)
		} else if exist > 0 && exist != int64(p.EditStateParameters.Header.StateID) {
			return fmt.Errorf(`State %s already exists`, p.EditStateParameters.Header.StateID)
		}
	}
	if err := p.AccessRights(p.EditStateParameters.Name, true); err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *EditStateParametersParser) Action() error {
	_, _, err := p.selectiveLoggingAndUpd([]string{"value", "conditions"}, []interface{}{p.EditStateParameters.Value,
		p.EditStateParameters.Conditions}, converter.Int64ToStr(p.EditStateParameters.Header.StateID)+"_state_parameters", []string{"name"},
		[]string{p.EditStateParameters.Name}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *EditStateParametersParser) Rollback() error {
	return p.autoRollback()
}

func (p EditStateParametersParser) Header() *tx.Header {
	return &p.EditStateParameters.Header
}
