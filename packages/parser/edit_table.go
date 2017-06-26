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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type EditTableParser struct {
	*Parser
	EditTable *tx.EditTable
}

func (p *EditTableParser) Init() error {
	editTable := &tx.EditTable{}
	if err := msgpack.Unmarshal(p.TxBinaryData, editTable); err != nil {
		return p.ErrInfo(err)
	}
	p.EditTable = editTable
	return nil
}

func (p *EditTableParser) Validate() error {
	err := p.generalCheck(`edit_table`, &p.EditTable.Header)
	if err != nil {
		return p.ErrInfo(err)
	}

	s := strings.Split(p.EditTable.Name, "_")
	if len(s) < 2 {
		return p.ErrInfo("incorrect table name")
	}
	prefix := s[0]
	if prefix != "global" && prefix != utils.Int64ToStr(p.EditTable.Header.StateID) {
		return p.ErrInfo("incorrect table name")
	}

	table := prefix + `_tables`
	exists, err := p.Single(`select count(*) from "`+table+`" where name = ?`, p.EditTable.Name).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if exists == 0 {
		return p.ErrInfo(`not exists`)
	}

	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.EditTable.ForSign(), p.EditTable.BinSignatures, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if err = p.AccessTable(p.EditTable.Name, `general_update`); err != nil {
		if err = p.AccessRights(`changing_tables`, false); err != nil {
			return p.ErrInfo(err)
		}
	}

	return nil
}

func (p *EditTableParser) Action() error {
	s := strings.Split(p.EditTable.Name, "_")
	if len(s) < 2 {
		return p.ErrInfo("incorrect table name")
	}
	prefix := s[0]
	if prefix != "global" && prefix != utils.Int64ToStr(p.EditTable.Header.StateID) {
		return p.ErrInfo("incorrect table name")
	}

	table := prefix + `_tables`
	tblname := p.EditTable.Name
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
	rbId, err := p.ExecSqlGetLastInsertId("INSERT INTO rollback ( data, block_id ) VALUES ( ?, ? )", "rollback", string(jsonData), p.BlockData.BlockId)
	if err != nil {
		return err
	}
	actions := map[string]string{
		"general_update": p.EditTable.GeneralUpdate,
		"new_column":     p.EditTable.NewColumn,
		"insert":         p.EditTable.Insert,
	}
	for _, action := range []string{"general_update", "new_column", "insert"} {
		if len(actions[action]) == 0 {
			return fmt.Errorf(`Parameter "%s" cannot be empty`, action)
		}
		if err := smart.CompileEval(actions[action], uint32(p.EditTable.Header.StateID)); err != nil {
			return err
		}
		actions[action] = strings.Replace(actions[action], `"`, `\"`, -1)
		err = p.ExecSql(`UPDATE "`+table+`" SET columns_and_permissions = jsonb_set(columns_and_permissions, '{`+action+`}', ?, true), rb_id = ? WHERE name = ?`,
			`"`+actions[action]+`"`, rbId, tblname)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	err = p.ExecSql("INSERT INTO rollback_tx ( block_id, tx_hash, table_name, table_id ) VALUES (?, [hex], ?, ?)", p.BlockData.BlockId, p.TxHash, table, p.EditTable.Name)
	if err != nil {
		return err
	}

	return nil
}

func (p *EditTableParser) Rollback() error {
	err := p.autoRollback()
	if err != nil {
		return err
	}
	return nil
}

func (p EditTableParser) Header() *tx.Header {
	return &p.EditTable.Header
}
