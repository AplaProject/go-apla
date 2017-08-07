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
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/sql"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type RestoreAccessParser struct {
	*Parser
	RestoreAccess *tx.RestoreAccess
}

func (p *RestoreAccessParser) Init() error {
	restoreAccess := &tx.RestoreAccess{}
	if err := msgpack.Unmarshal(p.TxBinaryData, restoreAccess); err != nil {
		return p.ErrInfo(err)
	}
	p.RestoreAccess = restoreAccess
	return nil
}

func (p *RestoreAccessParser) Validate() error {
	err := p.generalCheck(`system_restore_access`, &p.RestoreAccess.Header, map[string]string{})
	if err != nil {
		return p.ErrInfo(err)
	}

	// Check InputData
	verifyData := map[string][]interface{}{"int64": []interface{}{p.RestoreAccess.StateID}}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	if p.TxWalletID != sql.SysInt64(sql.RecoveryAddress) {
		return p.ErrInfo("p.TxWalletID != sql.RecoveryAddress")
	}

	restoreAccess := &model.SystemRestoreAccess{}
	err = restoreAccess.Get(p.RestoreAccess.StateID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if restoreAccess.Active == 0 {
		return p.ErrInfo("active = 0")
	}
	if restoreAccess.Close == 1 {
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
	if txTime-restoreAccess.Time < consts.CHANGE_KEY_PERIOD {
		return p.ErrInfo("CHANGE_KEY_PERIOD")
	}

	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.RestoreAccess.ForSign(), p.RestoreAccess.BinSignatures, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *RestoreAccessParser) Action() error {
	restoreAccess := &model.SystemRestoreAccess{}
	err := restoreAccess.Get(p.RestoreAccess.StateID)
	if err != nil {
		return p.ErrInfo(err)
	}
	value := `$citizen=` + converter.Int64ToStr(restoreAccess.CitizenID)
	_, _, err = p.selectiveLoggingAndUpd([]string{"value", "conditions"}, []interface{}{value, value}, p.TxStateIDStr+"_state_parameters", []string{"name"}, []string{"changing_tables"}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	_, _, err = p.selectiveLoggingAndUpd([]string{"value", "conditions"}, []interface{}{value, value}, p.TxStateIDStr+"_state_parameters", []string{"name"}, []string{"changing_smart_contracts"}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	_, _, err = p.selectiveLoggingAndUpd([]string{"close"}, []interface{}{"1"}, "system_restore_access", []string{"state_id"}, []string{converter.Int64ToStr(p.RestoreAccess.StateID)}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *RestoreAccessParser) Rollback() error {
	return p.autoRollback()
}

func (p *RestoreAccessParser) Header() *tx.Header {
	return &p.RestoreAccess.Header
}
