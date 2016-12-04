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
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

func (p *Parser) RestoreAccessRequestInit() error {

	fields := []map[string]string{{"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) RestoreAccessRequestFront() error {
	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	EGSPrice, err := p.getEGSPrice(`system_restore_access_request`)
	if err != nil {
		return p.ErrInfo(err)
	}

	// есть ли нужная сумма на кошельке
	err = p.checkSenderDLT(EGSPrice, 0)
	if err != nil {
		return p.ErrInfo(err)
	}

	data, err := p.OneRow("SELECT * FROM system_restore_access WHERE state_id  =  ?", p.TxStateID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(data) == 0 {
		return p.ErrInfo("!system_restore_access")
	}
	if data["active"] == 0 {
		return p.ErrInfo("active=0")
	}

	forSign := fmt.Sprintf("%s,%s,%d,%d", p.TxMap["type"], p.TxMap["time"], p.TxCitizenID, p.TxStateID)
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
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

func (p *Parser) RestoreAccessRequest() error {
	_, err := p.selectiveLoggingAndUpd([]string{"time", "close", "citizen_id"}, []interface{}{p.BlockData.Time, "0", p.TxCitizenID}, "system_restore_access", []string{"state_id"}, []string{utils.UInt32ToStr(p.TxStateID)}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) RestoreAccessRequestRollback() error {
	return p.autoRollback()
}