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
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	"gopkg.in/vmihailenco/msgpack.v2"
)

/*
Adding state tables should be spelled out in state settings
*/

type NewStateParametersParser struct {
	*Parser
	NewStateParameters *tx.NewStateParameters
}

func (p *NewStateParametersParser) Init() error {
	newStateParameters := &tx.NewStateParameters{}
	if err := msgpack.Unmarshal(p.TxBinaryData, newStateParameters); err != nil {
		return p.ErrInfo(err)
	}
	p.NewStateParameters = newStateParameters
	return nil
}

func (p *NewStateParametersParser) Validate() error {
	err := p.generalCheck(`new_state_parameters`, &p.NewStateParameters.Header, map[string]string{"conditions": p.NewStateParameters.Conditions})
	if err != nil {
		return p.ErrInfo(err)
	}

	if len(p.NewStateParameters.Conditions) > 0 {
		if err := smart.CompileEval(string(p.NewStateParameters.Conditions), uint32(p.NewStateParameters.Header.StateID)); err != nil {
			return p.ErrInfo(err)
		}
	}
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.NewStateParameters.ForSign(), p.NewStateParameters.BinSignatures, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *NewStateParametersParser) Action() error {
	txStateIDStr := converter.Int64ToStr(p.NewStateParameters.Header.StateID)
	_, err := p.selectiveLoggingAndUpd([]string{"name", "value", "conditions"}, []interface{}{p.NewStateParameters.Name, p.NewStateParameters.Value, p.NewStateParameters.Conditions}, txStateIDStr+"_state_parameters", nil, nil, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *NewStateParametersParser) Rollback() error {
	return p.autoRollback()
}

func (p NewStateParametersParser) Header() *tx.Header {
	return &p.NewStateParameters.Header
}
