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

type RestoreAccessRequestParser struct {
	*Parser
	RestoreAccessRequest *tx.RestoreAccessRequest
}

func (p *RestoreAccessRequestParser) Init() error {
	restoreAccessRequest := &tx.RestoreAccessRequest{}
	if err := msgpack.Unmarshal(p.TxBinaryData, restoreAccessRequest); err != nil {
		return p.ErrInfo(err)
	}
	p.RestoreAccessRequest = restoreAccessRequest
	return nil
}

func (p *RestoreAccessRequestParser) Validate() error {
	err := p.generalCheck(`system_restore_access_request`, &p.RestoreAccessRequest.Header)
	if err != nil {
		return p.ErrInfo(err)
	}

	data, err := p.OneRow("SELECT * FROM system_restore_access WHERE state_id  =  ?", p.RestoreAccessRequest.Header.StateID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(data) == 0 {
		return p.ErrInfo("!system_restore_access")
	}
	if data["active"] == 0 {
		return p.ErrInfo("active=0")
	}

	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.RestoreAccessRequest.ForSign(), p.RestoreAccessRequest.BinSignatures, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if err = p.AccessRights(`restore_access_condition`, false); err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *RestoreAccessRequestParser) Action() error {
	_, err := p.selectiveLoggingAndUpd([]string{"time", "close", "citizen_id"}, []interface{}{p.BlockData.Time, "0", p.RestoreAccessRequest.Header.UserID}, "system_restore_access", []string{"state_id"}, []string{utils.Int64ToStr(p.RestoreAccessRequest.StateID)}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *RestoreAccessRequestParser) Rollback() error {
	return p.autoRollback()
}

func (p RestoreAccessRequestParser) Header() *tx.Header {
	return &p.RestoreAccessRequest.Header
}
