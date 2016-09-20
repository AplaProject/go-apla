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
	"github.com/DayLightProject/go-daylight/packages/script"
)

/*
Adding state tables should be spelled out in state settings
*/

func (p *Parser) ChangeStateSettingsConditionsInit() error {

	fields := []map[string]string{{"state_id": "int64"}, {"parameter": "string"}, {"value": "string"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}



func (p *Parser) ChangeStateSettingsConditionsFront() error {
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
	conditions, err := p.Single(`SELECT change FROM `+p.States[p.TxMaps.Int64["state_id"]]+`_state_parameters WHERE parameter = ?`, p.TxMaps.String["parameter"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}

	vars := map[string]interface{}{
		`citizenId`: 	p.TxCitizenID,
		`walletId`: 	p.TxWalletID,
		`Table`:     	p.MyTable,
	}
	out, err := script.EvalIf(conditions, &vars)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !out {
		return p.ErrInfo("conditions false")
	}

	// Checking new condition
	vars = map[string]interface{}{
		`citizenId`: 	p.TxCitizenID,
		`walletId`: 	p.TxWalletID,
		`Table`:     	p.MyTableChecking,
	}
	out, err = script.EvalIf(p.TxMaps.String["conditions"], &vars)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !out {
		return p.ErrInfo("conditions false")
	}
	
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

func (p *Parser) ChangeStateSettingsConditions() error {
	err := p.selectiveLoggingAndUpd([]string{"value"}, []interface{}{p.TxMaps.String["value"]}, p.States[p.TxMaps.Int64["state_id"]]+"_state_parameters", []string{"parameter"}, []string{p.TxMaps.String["parameter"]}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangeStateSettingsConditionsRollback() error {
	return p.autoRollback()
}

func (p *Parser) ChangeStateSettingsConditionsRollbackFront() error {

	return nil
}
