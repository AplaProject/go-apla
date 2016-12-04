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
	"time"
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
)

func (p *Parser) RestoreAccessInit() error {

	fields := []map[string]string{{"state_id": "int64"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) RestoreAccessFront() error {
	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// Check InputData
	verifyData := map[string]string{"state_id": "int64"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
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
		txTime = p.BlockData.Time
	} else {
		txTime = time.Now().Unix() - 30 // просто на всякий случай небольшой запас
	}
	// прошел ли месяц с момента, когда кто-то запросил смену ключа
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

func (p *Parser) RestoreAccess() error {
	citizen_id, err := p.Single(`SELECT citizen_id FROM system_restore_access WHERE state_id = ?`, p.TxMaps.Int64["state_id"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	value := `$citizen=`+citizen_id
	_, err = p.selectiveLoggingAndUpd([]string{"value", "conditions"}, []interface{}{value, value}, p.TxStateIDStr+"_state_parameters", []string{"name"}, []string{"changing_tables"}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	_, err = p.selectiveLoggingAndUpd([]string{"value", "conditions"}, []interface{}{value, value}, p.TxStateIDStr+"_state_parameters", []string{"name"}, []string{"changing_smart_contracts"}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	_, err = p.selectiveLoggingAndUpd([]string{"close"}, []interface{}{"1"}, "system_restore_access", []string{"state_id"}, []string{utils.Int64ToStr(p.TxMaps.Int64["state_id"])}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) RestoreAccessRollback() error {
	return p.autoRollback()
}
