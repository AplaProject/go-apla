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
	"encoding/json"
)

func (p *Parser) NewCitizenInit() error {

	fields := []map[string]string{{"public_key": "bytes"}, {"state_id": "int64"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	p.TxMap["public_key_hex"] = utils.BinToHex(p.TxMap["public_key"])
	p.TxMaps.Bytes["public_key_hex"] = utils.BinToHex(p.TxMaps.Bytes["public_key"])
	return nil
}

func (p *Parser) NewCitizenFront() error {
	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// To not record too small or too big key
	if !utils.CheckInputData(p.TxMap["public_key_hex"], "public_key") {
		return utils.ErrInfo(fmt.Errorf("incorrect public_key %s", p.TxMap["public_key_hex"]))
	}

	// We get a set of custom fields that need to be in the tx
	additionalFields, err := p.Single(`SELECT fields FROM citizen_fields WHERE state_id = ?`, p.TxMaps.Int64["state_id"]).Bytes()
	if err != nil {
		return p.ErrInfo(err)
	}

	additionalFieldsMap := []map[string]string{}
	err = json.Unmarshal(additionalFields, &additionalFieldsMap)
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := make(map[string]string)
	for _, date := range additionalFieldsMap {
		verifyData[date["name"]] = date["txType"]
	}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// Citizens can only add a citizen of the same country

	// One who adds a citizen must be a valid representative body appointed in ds_state_settings


	// must be supplemented
	forSign := fmt.Sprintf("%s,%s,%d", p.TxMap["type"], p.TxMap["time"], p.TxWalletID)
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) NewCitizen() error {

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

func (p *Parser) NewCitizenRollback() error {
	return p.autoRollback()
}

func (p *Parser) NewCitizenRollbackFront() error {

	return nil

}
