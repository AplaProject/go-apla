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
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// NewColumnInit initializes NewColumn transaction
func (p *Parser) NewColumnInit() error {

	fields := []map[string]string{{"table_name": "string"}, {"column_name": "string"}, {"permissions": "string"}, {"index": "int64"}, {"column_type": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

// NewColumnFront checks conditions of NewColumn transaction
func (p *Parser) NewColumnFront() error {
	err := p.generalCheck(`new_column`)
	if err != nil {
		return p.ErrInfo(err)
	}

	// Check InputData
	verifyData := map[string]string{"table_name": "string", "column_name": "string", "permissions": "conditions", "index": "int64", "column_type": "column_type"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	prefix, err := utils.GetPrefix(p.TxMaps.String["table_name"], p.TxStateIDStr)
	if err != nil {
		return p.ErrInfo(err)
	}
	table := prefix + `_tables`
	exists, err := p.Single(`select count(*) from "`+table+`" where (columns_and_permissions->'update'-> ? ) is not null AND name = ?`, p.TxMaps.String["column_name"], p.TxMaps.String["table_name"]).Int64()
	log.Debug(`select count(*) from "`+table+`" where (columns_and_permissions->'update'-> ? ) is not null AND name = ?`, p.TxMaps.String["column_name"], p.TxMaps.String["table_name"])
	if err != nil {
		return p.ErrInfo(err)
	}
	if exists > 0 {
		return p.ErrInfo(`column exists`)
	}

	count, err := p.Single("SELECT count(column_name) FROM information_schema.columns WHERE table_name=?", p.TxMaps.String["table_name"]).Int64()
	if count >= consts.MAX_COLUMNS+2 /*id + rb_id*/ {
		return fmt.Errorf(`Too many columns. Limit is %d`, consts.MAX_COLUMNS)
	}
	if p.TxMaps.Int64["index"] > 0 {
		count, err := p.NumIndexes(p.TxMaps.String["table_name"])
		if err != nil {
			return p.ErrInfo(err)
		}
		if count >= consts.MAX_INDEXES {
			return fmt.Errorf(`Too many indexes. Limit is %d`, consts.MAX_INDEXES)
		}
	}

	forSign := fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxCitizenID, p.TxStateID, p.TxMap["table_name"], p.TxMap["column_name"], p.TxMap["permissions"], p.TxMap["index"], p.TxMap["column_type"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if err := p.AccessTable(p.TxMaps.String["table_name"], "new_column"); err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

// NewColumn proceeds NewColumn transaction
func (p *Parser) NewColumn() error {

	tblname := p.TxMaps.String["table_name"]
	prefix, err := utils.GetPrefix(tblname, p.TxStateIDStr)
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
	rbID, err := p.ExecSQLGetLastInsertID("INSERT INTO rollback ( data, block_id ) VALUES ( ?, ? )", "rollback", string(jsonData), p.BlockData.BlockId)
	if err != nil {
		return err
	}
	err = p.ExecSQL(`UPDATE "`+table+`" SET columns_and_permissions = jsonb_set(columns_and_permissions, '{update, `+p.TxMaps.String["column_name"]+`}', ?, true), rb_id = ? WHERE name = ?`, `"`+converter.EscapeForJSON(p.TxMaps.String["permissions"])+`"`, rbID, tblname)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSQL("INSERT INTO rollback_tx ( block_id, tx_hash, table_name, table_id ) VALUES (?, [hex], ?, ?)", p.BlockData.BlockId, p.TxHash, table, tblname)
	if err != nil {
		return err
	}

	colType := ``
	switch p.TxMaps.String["column_type"] {
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

	err = p.ExecSQL(`ALTER TABLE "` + tblname + `" ADD COLUMN ` + p.TxMaps.String["column_name"] + ` ` + colType)
	if err != nil {
		return err
	}

	if p.TxMaps.Int64["index"] == 1 {
		err = p.ExecSQL(`CREATE INDEX "` + tblname + `_` + p.TxMaps.String["column_name"] + `_index" ON "` + tblname + `" (` + p.TxMaps.String["column_name"] + `)`)
		if err != nil {
			return err
		}
	}

	return nil
}

// NewColumnRollback rollbacks NewColumn transaction
func (p *Parser) NewColumnRollback() error {
	err := p.autoRollback()
	if err != nil {
		return err
	}
	err = p.ExecSQL(`ALTER TABLE "` + p.TxMaps.String["table_name"] + `" DROP COLUMN ` + p.TxMaps.String["column_name"] + ``)
	if err != nil {
		return err
	}
	/*
		if p.TxMaps.Int64["index"] == 1 {
			err = p.ExecSQL(`DROP INDEX "` + p.TxMaps.String["table_name"] + `_` + p.TxMaps.String["column_name"] + `_index"`)
			if err != nil {
				return err
			}
		}*/
	return nil
}
