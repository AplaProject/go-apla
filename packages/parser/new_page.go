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
	"github.com/DayLightProject/go-daylight/packages/utils"
	"fmt"
)

func (p *Parser) NewPageInit() error {

	fields := []map[string]string{{"global": "string"},{"name": "string"}, {"value": "string"}, {"menu": "string"},  {"conditions": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewPageFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// Check InputData
	/*verifyData := map[string]string{"name": "string", "value": "string", "menu": "string", "conditions": "string"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}*/

	/*
	Check conditions
	...
	*/

	// must be supplemented
	forSign := fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxCitizenID, p.TxStateID, p.TxMap["global"],p.TxMap["name"], p.TxMap["value"], p.TxMap["menu"], p.TxMap["conditions"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) NewPage() error {

	prefix := p.TxStateIDStr
	if p.TxMaps.String["global"]=="1" {
		prefix = "global"
	}
	err := p.selectiveLoggingAndUpd([]string{"value", "menu", "conditions"}, []interface{}{p.TxMaps.String["value"], p.TxMaps.String["menu"], p.TxMaps.String["conditions"]}, prefix+"_pages", nil, nil, true)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) NewPageRollback() error {
	return p.autoRollback()
}

func (p *Parser) NewPageRollbackFront() error {
	return nil
}