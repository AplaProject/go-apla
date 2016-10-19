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

	"github.com/DayLightProject/go-daylight/packages/smart"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) EditContractInit() error {

	fields := []map[string]string{{"global": "int64"}, {"id": "string"}, {"value": "string"}, {"conditions": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) EditContractFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// Check the system limits. You can not send more than X time a day this TX
	// ...

	// Check InputData
	verifyData := map[string]string{}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// must be supplemented
	forSign := fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxCitizenID, p.TxStateID, p.TxMap["global"], p.TxMap["id"], p.TxMap["value"], p.TxMap["conditions"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if len(p.TxMap["conditions"]) > 0 {
		if err := smart.CompileEval(string(p.TxMap["conditions"])); err != nil {
			return p.ErrInfo(err)
		}
	}
	conditions, err := p.Single(`SELECT conditions FROM "`+utils.Int64ToStr(int64(p.TxStateID))+`_smart_contracts" WHERE id = ?`, p.TxMaps.String["id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(conditions) > 0 {
		ret, err := smart.EvalIf(conditions, &map[string]interface{}{`state`: p.TxStateID,
			`citizen`: p.TxCitizenID, `wallet`: p.TxWalletID})
		if err != nil {
			return p.ErrInfo(err)
		}
		if !ret {
			return fmt.Errorf(`Access denied`)
		}
	}

	return nil
}

func (p *Parser) EditContract() error {

	prefix := `global`
	if p.TxMaps.Int64["global"] == 0 {
		prefix = p.TxStateIDStr
	}
	root, err := smart.CompileBlock(p.TxMaps.String["value"])
	if err != nil {
		return p.ErrInfo(err)
	}

	_, err = p.selectiveLoggingAndUpd([]string{"value", "conditions"}, []interface{}{p.TxMaps.String["value"], p.TxMaps.String["conditions"]}, prefix+"_smart_contracts", []string{"id"}, []string{p.TxMaps.String["id"]}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	smart.FlushBlock(root)
	return nil
}

func (p *Parser) EditContractRollback() error {
	return p.autoRollback()
}

/*func (p *Parser) EditContractRollbackFront() error {
	return nil
}*/
