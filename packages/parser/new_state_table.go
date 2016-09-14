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

/*
Adding state tables should be spelled out in state settings
*/

func (p *Parser) NewStateTableInit() error {

	fields := []map[string]string{{"public_key": "bytes"}, {"table_name": "string"}, {"table_columns": "string"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewStateTableFront() error {
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

	// New state table can only add a citizen of the same country
	// ...


	// Check the condition that must be met to complete this transaction
	// select value from ds_state_settings where name = "new_state_table"
	// ...




	// must be supplemented
	forSign := fmt.Sprintf("%s,%s,%d", p.TxMap["type"], p.TxMap["time"], p.TxMap["state_id"], p.TxCitizenID)
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) NewStateTable() error {

	stateCode, err := p.Single(`SELECT state_code FROM states WHERE state_id = ?`, p.TxMaps.Int64["state_id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql(`INSERT INTO `+stateCode+`_citizens ( public_key_0, block_id ) VALUES ( [hex], ? )`, p.TxMap["public_key_hex"], p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewStateTableRollback() error {

	stateCode, err := p.Single(`SELECT state_code FROM states WHERE state_id = ?`, p.TxMaps.Int64["state_id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql(`DELETE FROM `+stateCode+`_citizens WHERE block_id = ?`, p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewStateTableRollbackFront() error {

	return nil

}
