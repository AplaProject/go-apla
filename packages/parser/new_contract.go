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
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

func (p *Parser) NewContractInit() error {

	fields := []map[string]string{{"global": "int64"}, {"name": "string"}, {"value": "string"}, {"conditions": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewContractFront() error {

	err := p.generalCheck(`new_contract`)
	if err != nil {
		return p.ErrInfo(err)
	}

	// Check the system limits. You can not send more than X time a day this TX
	// ...

	// Check InputData
	name := p.TxMaps.String["name"]
	if off := strings.IndexByte(name, '#'); off > 0 {
		p.TxMap["name"] = []byte(name[:off])
		p.TxMaps.String["name"] = name[:off]
		address := lib.StringToAddress(name[off+1:])
		if address == 0 {
			return p.ErrInfo(fmt.Errorf(`wrong wallet %s`, name[off+1:]))
		}
		p.TxMaps.Int64["wallet_contract"] = address
	}
	verifyData := map[string]string{"global": "int64", "name": "string"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// must be supplemented
	forSign := fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxCitizenID, p.TxStateID, p.TxMap["global"], name, p.TxMap["value"], p.TxMap["conditions"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	prefix := `global`
	if p.TxMaps.Int64["global"] == 0 {
		prefix = p.TxStateIDStr
	}
	if len(p.TxMap["conditions"]) > 0 {
		if err := smart.CompileEval(string(p.TxMap["conditions"]), uint32(p.TxStateID)); err != nil {
			return p.ErrInfo(err)
		}
	}

	if exist, err := p.Single(`select id from "`+prefix+"_smart_contracts"+`" where name=?`, p.TxMap["name"]).Int64(); err != nil {
		return p.ErrInfo(err)
	} else if exist > 0 {
		return p.ErrInfo(fmt.Sprintf("The contract %s already exists", p.TxMap["name"]))
	}
	return nil
}

func (p *Parser) NewContract() error {

	prefix := `global`
	if p.TxMaps.Int64["global"] == 0 {
		prefix = p.TxStateIDStr
	}
	var wallet int64
	if wallet = p.TxCitizenID; wallet == 0 {
		wallet = p.TxWalletID
	}
	root, err := smart.CompileBlock(p.TxMaps.String["value"], prefix, false, 0)
	if err != nil {
		return p.ErrInfo(err)
	}
	if val, ok := p.TxMaps.Int64["wallet_contract"]; ok {
		wallet = val
	}

	tblid, err := p.selectiveLoggingAndUpd([]string{"name", "value", "conditions", "wallet_id"},
		[]interface{}{p.TxMaps.String["name"], p.TxMaps.String["value"], p.TxMaps.String["conditions"],
			wallet}, prefix+"_smart_contracts", nil, nil, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	for i, item := range root.Children {
		if item.Type == script.OBJ_CONTRACT {
			root.Children[i].Info.(*script.ContractInfo).TblId = utils.StrToInt64(tblid)
		}
	}

	smart.FlushBlock(root)
	return nil
}

func (p *Parser) NewContractRollback() error {
	return p.autoRollback()
}

/*func (p *Parser) NewContractRollbackFront() error {
	return nil
}
*/
