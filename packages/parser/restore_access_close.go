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

type RestoreAccessCloseParser struct {
	*Parser
	RestoreAccessClose *tx.RestoreAccessClose
}

func (p *RestoreAccessCloseParser) Init() error {
	restoreAccessClose := &tx.RestoreAccessClose{}
	if err := msgpack.Unmarshal(p.TxBinaryData, restoreAccessClose); err != nil {
		return p.ErrInfo(err)
	}
	p.RestoreAccessClose = restoreAccessClose
	return nil
}

func (p *RestoreAccessCloseParser) Validate() error {
	err := p.generalCheck(`system_restore_access_close`, &p.RestoreAccessClose.Header)
	if err != nil {
		return p.ErrInfo(err)
	}

	// check whether or not already close
	close, err := p.Single("SELECT close FROM system_restore_access WHERE user_id  =  ? AND state_id = ?", p.RestoreAccessClose.Header.UserID, p.RestoreAccessClose.Header.StateID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if close > 0 {
		return p.ErrInfo("close=1")
	}

	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.RestoreAccessClose.ForSign(), p.TxMap["sign"], false)
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

func (p *RestoreAccessCloseParser) Action() error {
	_, err := p.selectiveLoggingAndUpd([]string{"close"}, []interface{}{"1"}, "system_restore_access", []string{"state_id"}, []string{utils.Int64ToStr(p.RestoreAccessClose.Header.StateID)}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *RestoreAccessCloseParser) Rollback() error {
	return p.autoRollback()
}

func (p *RestoreAccessCloseParser) Header() *tx.Header {
	return &p.RestoreAccessClose.Header
}
