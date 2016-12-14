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

func (p *Parser) EditLangInit() error {

	fields := []map[string]string{ /*{"global": "int64"},*/ {"name": "string"}, {"res": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) EditLangFront() error {

	err := p.generalCheck()
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
	forSign := fmt.Sprintf("%s,%s,%d,%d,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxCitizenID, p.TxStateID,
		/*p.TxMap["global"],*/ p.TxMap["name"], p.TxMap["res"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if err = p.AccessRights(`changing_language`, false); err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) EditLang() error {

	/*	prefix := `global`
		if p.TxMaps.Int64["global"] == 0 {
			prefix = p.TxStateIDStr
		}
	*/
	prefix := p.TxStateIDStr
	_, err := p.selectiveLoggingAndUpd([]string{"name", "res"}, []interface{}{p.TxMaps.String["name"],
		p.TxMaps.String["res"]}, prefix+"_languages", []string{"name"}, []string{p.TxMaps.String["name"]}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) EditLangRollback() error {
	return p.autoRollback()
}
