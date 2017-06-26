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

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type NewColumnParser struct {
	*Parser
	NewColumn *tx.NewColumn
}

func (p *NewColumnParser) Init() error {
	newColumn := &tx.NewColumn{}
	if err := msgpack.Unmarshal(p.TxBinaryData, newColumn); err != nil {
		return p.ErrInfo(err)
	}
	p.NewColumn = newColumn
	return nil
}

func (p *NewColumnParser) Validate() error {
	p.TxMap["permissions"] = []byte(p.NewColumn.Permissions)
	err := p.generalCheck(`new_column`, &p.NewColumn.Header)
	if err != nil {
		return p.ErrInfo(err)
	}

	// Check InputData
	verifyData := map[string][]interface{}{"string": []interface{}{p.NewColumn.TableName, p.NewColumn.ColumnName}, "conditions": []interface{}{p.NewColumn.Permissions}, "int64": []interface{}{p.NewColumn.Index}, "column_type": []interface{}{p.NewColumn.ColumnType}}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	prefix, err := utils.GetPrefix(p.NewColumn.TableName, utils.Int64ToStr(p.NewColumn.Header.StateID))
	if err != nil {
		return p.ErrInfo(err)
	}
	table := prefix + `_tables`
	exists, err := p.Single(`select count(*) from "`+table+`" where (columns_and_permissions->'update'-> ? ) is not null AND name = ?`, p.NewColumn.ColumnName, p.NewColumn.TableName).Int64()
	log.Debug(`select count(*) from "`+table+`" where (columns_and_permissions->'update'-> ? ) is not null AND name = ?`, p.NewColumn.ColumnName, p.NewColumn.TableName)
	if err != nil {
		return p.ErrInfo(err)
	}
	if exists > 0 {
		return p.ErrInfo(`column exists`)
	}

	count, err := p.Single("SELECT count(column_name) FROM information_schema.columns WHERE table_name=?", p.NewColumn.TableName).Int64()
	if count >= consts.MAX_COLUMNS+2 /*id + rb_id*/ {
		return fmt.Errorf(`Too many columns. Limit is %d`, consts.MAX_COLUMNS)
	}
	if utils.StrToInt64(p.NewColumn.Index) > 0 {
		count, err := p.NumIndexes(p.NewColumn.TableName)
		if err != nil {
			return p.ErrInfo(err)
		}
		if count >= consts.MAX_INDEXES {
			return fmt.Errorf(`Too many indexes. Limit is %d`, consts.MAX_INDEXES)
		}
	}

	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.NewColumn.ForSign(), p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if err := p.AccessTable(p.NewColumn.TableName, "new_column"); err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *NewColumnParser) Action() error {
	tblname := p.NewColumn.TableName
	prefix, err := utils.GetPrefix(tblname, utils.Int64ToStr(p.NewColumn.StateID))
	if err != nil {
		return p.ErrInfo(err)
	}
	table := prefix + `_tables`
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
	err = p.ExecSql(`UPDATE "`+table+`" SET columns_and_permissions = jsonb_set(columns_and_permissions, '{update, `+p.NewColumn.ColumnName+`}', ?, true), rb_id = ? WHERE name = ?`, `"`+lib.EscapeForJSON(p.NewColumn.Permissions)+`"`, rbId, tblname)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql("INSERT INTO rollback_tx ( block_id, tx_hash, table_name, table_id ) VALUES (?, [hex], ?, ?)", p.BlockData.BlockId, p.TxHash, table, tblname)
	if err != nil {
		return err
	}

	colType := ``
	switch p.NewColumn.ColumnType {
	case "text":
		colType = `varchar(102400)`
	case "int64":
		colType = `bigint`
	case "time":
		colType = `timestamp`
	case "hash":
		colType = `bytea`
	case "money":
		colType = `decimal(30,0)`
	case "double":
		colType = `double precision`
	}

	err = p.ExecSql(`ALTER TABLE "` + tblname + `" ADD COLUMN ` + p.NewColumn.ColumnName + ` ` + colType)
	if err != nil {
		return err
	}

	if p.NewColumn.Index == "1" {
		err = p.ExecSql(`CREATE INDEX "` + tblname + `_` + p.NewColumn.ColumnName + `_index" ON "` + tblname + `" (` + p.NewColumn.ColumnName + `)`)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *NewColumnParser) Rollback() error {
	err := p.autoRollback()
	if err != nil {
		return err
	}
	err = p.ExecSql(`ALTER TABLE "` + p.NewColumn.TableName + `" DROP COLUMN ` + p.NewColumn.ColumnName + ``)
	if err != nil {
		return err
	}
	return nil
}

func (p NewColumnParser) Header() *tx.Header {
	return &p.NewColumn.Header
}
