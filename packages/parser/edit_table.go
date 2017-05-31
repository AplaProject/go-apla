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
	//"encoding/json"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// EditTableInit initializes EditTable transaction
func (p *Parser) EditTableInit() error {

	fields := []map[string]string{{"table_name": "string"}, {"general_update": "string"}, {"insert": "string"}, {"new_column": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

// EditTableFront checks conditions of EditTable transaction
func (p *Parser) EditTableFront() error {
	err := p.generalCheck(`edit_table`)
	if err != nil {
		return p.ErrInfo(err)
	}

	s := strings.Split(p.TxMaps.String["table_name"], "_")
	if len(s) < 2 {
		return p.ErrInfo("incorrect table name")
	}
	prefix := s[0]
	if prefix != "global" && prefix != p.TxStateIDStr {
		return p.ErrInfo("incorrect table name")
	}

	// Check InputData
	/*verifyData := map[string]string{"table_name": "string", "column_name": "word", "permissions": "string"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}*/

	table := prefix + `_tables`
	exists, err := p.Single(`select count(*) from "`+table+`" where name = ?`, p.TxMaps.String["table_name"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if exists == 0 {
		return p.ErrInfo(`not exists`)
	}

	forSign := fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxCitizenID, p.TxStateID, p.TxMap["table_name"], p.TxMap["general_update"], p.TxMap["insert"], p.TxMap["new_column"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if err = p.AccessTable(p.TxMaps.String["table_name"], `general_update`); err != nil {
		if err = p.AccessRights(`changing_tables`, false); err != nil {
			return p.ErrInfo(err)
		}
	}

	return nil
}

// EditTable proceeds EditTable transaction
func (p *Parser) EditTable() error {

	s := strings.Split(p.TxMaps.String["table_name"], "_")
	if len(s) < 2 {
		return p.ErrInfo("incorrect table name")
	}
	prefix := s[0]
	if prefix != "global" && prefix != p.TxStateIDStr {
		return p.ErrInfo("incorrect table name")
	}

	table := prefix + `_tables`
	tblname := p.TxMaps.String["table_name"]
	logData, err := p.OneRow(`SELECT columns_and_permissions, rb_id FROM "`+table+`" where name=?`, tblname).String()
	if err != nil {
		return err
	}
	jsonMap := make(map[string]string)
	for k, v := range logData {
		if k == p.AllPkeys[table] {
			continue
		}
		jsonMap[k] = v
		if k == "rb_id" {
			k = "prev_rb_id"
		}
	}
	jsonData, _ := json.Marshal(jsonMap)
	if err != nil {
		return err
	}
	rbID, err := p.ExecSQLGetLastInsertID("INSERT INTO rollback ( data, block_id ) VALUES ( ?, ? )", "rollback", string(jsonData), p.BlockData.BlockId)
	if err != nil {
		return err
	}
	for _, action := range []string{"general_update", "new_column", "insert"} {
		if len(p.TxMaps.String[action]) == 0 {
			return fmt.Errorf(`Parameter "%s" cannot be empty`, action)
		}
		if err := smart.CompileEval(p.TxMaps.String[action], uint32(p.TxStateID)); err != nil {
			return err
		}
		p.TxMaps.String[action] = strings.Replace(p.TxMaps.String[action], `"`, `\"`, -1)
		err = p.ExecSQL(`UPDATE "`+table+`" SET columns_and_permissions = jsonb_set(columns_and_permissions, '{`+action+`}', ?, true), rb_id = ? WHERE name = ?`,
			`"`+p.TxMaps.String[action]+`"`, rbID, tblname)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	/*	err = p.ExecSQL(`UPDATE "`+table+`" SET columns_and_permissions = jsonb_set(columns_and_permissions, '{general_update}', ?, true), rb_id = ? WHERE name = ?`, `"`+p.TxMaps.String["general_update"]+`"`, rbID, p.TxMaps.String["table_name"])
		if err != nil {
			return p.ErrInfo(err)
		}*/

	err = p.ExecSQL("INSERT INTO rollback_tx ( block_id, tx_hash, table_name, table_id ) VALUES (?, [hex], ?, ?)", p.BlockData.BlockId, p.TxHash, table, p.TxMaps.String["table_name"])
	if err != nil {
		return err
	}

	return nil
}

// EditTableRollback rollbacks EditTable transaction
func (p *Parser) EditTableRollback() error {
	err := p.autoRollback()
	if err != nil {
		return err
	}
	return nil
}
