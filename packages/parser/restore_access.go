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
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// RestoreAccessInit initializes RestoreAccess transaction
func (p *Parser) RestoreAccessInit() error {

	fields := []map[string]string{{"state_id": "int64"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

// RestoreAccessFront checks conditions of RestoreAccess transaction
func (p *Parser) RestoreAccessFront() error {
	err := p.generalCheck(`system_restore_access`)
	if err != nil {
		return p.ErrInfo(err)
	}

	// Check InputData
	verifyData := map[string]string{"state_id": "int64"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	if p.TxWalletID != consts.RECOVERY_ADDRESS {
		return p.ErrInfo("p.TxWalletID != consts.RECOVERY_ADDRESS")
	}

	data, err := p.OneRow("SELECT * FROM system_restore_access WHERE state_id  =  ?", p.TxMaps.Int64["state_id"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(data) == 0 {
		return p.ErrInfo("incorrect system_restore_access")
	}
	if data["active"] == 0 {
		return p.ErrInfo("active = 0")
	}
	if data["close"] == 1 {
		return p.ErrInfo("close = 1")
	}

	var txTime int64
	if p.BlockData != nil { // тр-ия пришла в блоке
		// transaction has come in the block
		txTime = p.BlockData.Time
	} else {
		txTime = time.Now().Unix() - 30 // просто на всякий случай небольшой запас
		// a small supply just in case
	}
	// прошел ли месяц с момента, когда кто-то запросил смену ключа
	// whether the month passed from the moment when someone requested changing of a key
	if txTime-data["change_key_time"] < consts.CHANGE_KEY_PERIOD {
		return p.ErrInfo("CHANGE_KEY_PERIOD")
	}

	forSign := fmt.Sprintf("%s,%s,%d,%d,%s", p.TxMap["type"], p.TxMap["time"], p.TxCitizenID, p.TxStateID, p.TxMap["state_id"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

// RestoreAccess proceeds RestoreAccess transaction
func (p *Parser) RestoreAccess() error {
	citizenID, err := p.Single(`SELECT citizen_id FROM system_restore_access WHERE state_id = ?`, p.TxMaps.Int64["state_id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	value := `$citizen=` + citizenID
	_, err = p.selectiveLoggingAndUpd([]string{"value", "conditions"}, []interface{}{value, value}, p.TxStateIDStr+"_state_parameters", []string{"name"}, []string{"changing_tables"}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	_, err = p.selectiveLoggingAndUpd([]string{"value", "conditions"}, []interface{}{value, value}, p.TxStateIDStr+"_state_parameters", []string{"name"}, []string{"changing_smart_contracts"}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	_, err = p.selectiveLoggingAndUpd([]string{"close"}, []interface{}{"1"}, "system_restore_access", []string{"state_id"}, []string{converter.Int64ToStr(p.TxMaps.Int64["state_id"])}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

// RestoreAccessRollback rollbacks RestoreAccess transaction
func (p *Parser) RestoreAccessRollback() error {
	return p.autoRollback()
}
