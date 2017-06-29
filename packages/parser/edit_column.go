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
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type EditColumnParser struct {
	*Parser
	EditColumn *tx.EditColumn
}

func (p *EditColumnParser) Init() error {
	editColumn := &tx.EditColumn{}
	if err := msgpack.Unmarshal(p.TxBinaryData, editColumn); err != nil {
		return p.ErrInfo(err)
	}
	p.EditColumn = editColumn
	return nil
}

func (p *EditColumnParser) Validate() error {
	err := p.generalCheck(`edit_column`, &p.EditColumn.Header, map[string]string{"permissions": p.EditColumn.Permissions})
	if err != nil {
		return p.ErrInfo(err)
	}

	stateIdStr := utils.Int64ToStr(p.EditColumn.Header.StateID)
	table := stateIdStr + `_tables`
	if strings.HasPrefix(stateIdStr, `global`) {
		table = `global_tables`
	}
	exists, err := p.Single(`select count(*) from "`+table+`" where (columns_and_permissions->'update'-> ? ) is not null AND name = ?`, p.EditColumn.ColumnName, p.EditColumn.TableName).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if exists == 0 {
		return p.ErrInfo(`column not exists`)
	}

	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.EditColumn.ForSign(), p.EditColumn.BinSignatures, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if err = p.AccessTable(p.EditColumn.TableName, `general_update`); err != nil {
		return err
	}

	return nil
}

func (p *EditColumnParser) Action() error {
	stateIdStr := utils.Int64ToStr(p.EditColumn.Header.StateID)
	table := stateIdStr + `_tables`
	if strings.HasPrefix(p.EditColumn.TableName, `global`) {
		table = `global_tables`
	}
	logData, err := p.OneRow(`SELECT columns_and_permissions, rb_id FROM "`+table+`" where (columns_and_permissions->'update'-> ? ) is not null AND name = ?`,
		p.EditColumn.ColumnName, p.EditColumn.TableName).String()
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
	err = p.ExecSql(`UPDATE "`+table+`" SET columns_and_permissions = jsonb_set(columns_and_permissions, '{update, `+p.EditColumn.ColumnName+`}', ?, true), rb_id = ? WHERE name = ?`,
		`"`+lib.EscapeForJSON(p.EditColumn.Permissions)+`"`, rbId, p.EditColumn.TableName)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql("INSERT INTO rollback_tx ( block_id, tx_hash, table_name, table_id ) VALUES (?, [hex], ?, ?)", p.BlockData.BlockId, p.TxHash, table, p.EditColumn.TableName)
	if err != nil {
		return err
	}

	return nil
}

func (p *EditColumnParser) Rollback() error {
	err := p.autoRollback()
	if err != nil {
		return err
	}
	return nil
}

func (p EditColumnParser) Header() *tx.Header {
	return &p.EditColumn.Header
}
