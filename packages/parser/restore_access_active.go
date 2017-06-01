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

type RestoreAccessActiveParser struct {
	*Parser
}

func (p *RestoreAccessActiveParser) Init() error {
	fields := []map[string]string{{"secret": "bytes"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	p.TxMaps.String["secret_hex"] = string(utils.BinToHex(p.TxMaps.Bytes["secret"]))
	if p.TxMaps.String["secret_hex"] == "30" {
		p.TxMaps.Int64["active"] = 0
	} else {
		p.TxMaps.Int64["active"] = 1
	}
	p.TxMaps.String["secret_hex"] = string(utils.BinToHex(p.TxMaps.Bytes["secret"]))

	return nil
}

func (p *RestoreAccessActiveParser) Validate() error {
	err := p.generalCheck(`system_restore_access_active`)
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(p.TxMaps.Bytes["secret"]) > 2048 {
		return p.ErrInfo("len secret > 2048")
	}

	// check that there is no repeat shift
	active, err := p.Single("SELECT active FROM system_restore_access WHERE state_id = ?", p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if active == p.TxMaps.Int64["active"] {
		return p.ErrInfo("active")
	}
	forSign := fmt.Sprintf("%s,%s,%d,%d,%s", p.TxMap["type"], p.TxMap["time"], p.TxCitizenID, p.TxStateID, p.TxMap["secret_hex"])
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

func (p *RestoreAccessActiveParser) Action() error {
	_, err := p.selectiveLoggingAndUpd([]string{"active", "secret"}, []interface{}{p.TxMaps.Int64["active"], p.TxMaps.Bytes["secret"]}, "system_restore_access", []string{"state_id"}, []string{utils.UInt32ToStr(p.TxStateID)}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *RestoreAccessActiveParser) Rollback() error {
	return p.autoRollback()
}
