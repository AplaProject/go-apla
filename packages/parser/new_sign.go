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
	//	"encoding/json"
	"fmt"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// NewSignInit initializes NewSign transaction
func (p *Parser) NewSignInit() error {

	fields := []map[string]string{{"global": "int64"}, {"name": "string"}, {"value": "string"}, {"conditions": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

// NewSignFront checks conditions of NewSign transaction
func (p *Parser) NewSignFront() error {

	err := p.generalCheck(`new_sign`)
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
	forSign := fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxCitizenID, p.TxStateID,
		p.TxMap["global"], p.TxMap["name"], p.TxMap["value"], p.TxMap["conditions"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if err = p.AccessRights(`changing_signature`, false); err != nil {
		return p.ErrInfo(err)
	}
	prefix := `global`
	if p.TxMaps.Int64["global"] == 0 {
		prefix = p.TxStateIDStr
	}
	if exist, err := p.Single(`select name from "`+prefix+"_signatures"+`" where name=?`, p.TxMap["name"]).String(); err != nil {
		return p.ErrInfo(err)
	} else if len(exist) > 0 {
		return p.ErrInfo(fmt.Sprintf("The signature %s already exists", p.TxMap["name"]))
	}
	return nil
}

// NewSign proceeds NewSign transaction
func (p *Parser) NewSign() error {

	prefix := `global`
	if p.TxMaps.Int64["global"] == 0 {
		prefix = p.TxStateIDStr
	}
	_, err := p.selectiveLoggingAndUpd([]string{"name", "value", "conditions"}, []interface{}{p.TxMaps.String["name"], p.TxMaps.String["value"], p.TxMaps.String["conditions"]}, prefix+"_signatures", nil, nil, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

// NewSignRollback rollbacks NewSign transaction
func (p *Parser) NewSignRollback() error {
	return p.autoRollback()
}
