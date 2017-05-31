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

	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// EditSignInit initializes EditSign transaction
func (p *Parser) EditSignInit() error {

	fields := []map[string]string{{"global": "int64"}, {"name": "string"}, {"value": "string"}, {"conditions": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

// EditSignFront checks conditions of EditSign transaction
func (p *Parser) EditSignFront() error {

	err := p.generalCheck(`edit_sign`)
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
	forSign := fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxCitizenID, p.TxStateID, p.TxMap["global"], p.TxMap["name"], p.TxMap["value"], p.TxMap["conditions"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	prefix := utils.Int64ToStr(int64(p.TxStateID))
	if len(p.TxMap["conditions"]) > 0 {
		if err := smart.CompileEval(string(p.TxMap["conditions"]), uint32(p.TxStateID)); err != nil {
			return p.ErrInfo(err)
		}
	}
	conditions, err := p.Single(`SELECT conditions FROM "`+prefix+`_signatures" WHERE name = ?`, p.TxMaps.String["name"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(conditions) > 0 {
		ret, err := p.EvalIf(conditions)
		if err != nil {
			return p.ErrInfo(err)
		}
		if !ret {
			if err = p.AccessRights(`changing_signatures`, false); err != nil {
				return p.ErrInfo(err)
			}
		}
	}

	return nil
}

// EditSign proceeds EditSign transaction
func (p *Parser) EditSign() error {

	prefix := `global`
	if p.TxMaps.Int64["global"] == 0 {
		prefix = p.TxStateIDStr
	}
	_, err := p.selectiveLoggingAndUpd([]string{"value", "conditions"}, []interface{}{p.TxMaps.String["value"],
		p.TxMaps.String["conditions"]}, prefix+"_signatures", []string{"name"}, []string{p.TxMaps.String["name"]}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

// EditSignRollback rollbacks EditSign transaction
func (p *Parser) EditSignRollback() error {
	return p.autoRollback()
}
