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
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

/*
Adding state tables should be spelled out in state settings
*/

// NewTableInit initializes NewTable transaction
func (p *Parser) NewTableInit() error {

	fields := []map[string]string{{"global": "int64"}, {"table_name": "string"}, {"columns": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

// NewTableFront checks conditions of NewTable transaction
func (p *Parser) NewTableFront() error {

	err := p.generalCheck(`add_table`)
	if err != nil {
		return p.ErrInfo(err)
	}

	// Check the system limits. You can not send more than X time a day this TX
	// ...

	// Check InputData
	verifyData := map[string]string{"global": "int64", "table_name": "string"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	var cols [][]string
	err = json.Unmarshal([]byte(p.TxMaps.String["columns"]), &cols)
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(cols) == 0 {
		return p.ErrInfo(`len(cols) == 0`)
	}
	if len(cols) > consts.MAX_COLUMNS {
		return fmt.Errorf(`Too many columns. Limit is %d`, consts.MAX_COLUMNS)
	}
	var indexes int
	for _, data := range cols {
		if len(data) != 3 {
			return p.ErrInfo(`len(data)!=3`)
		}
		if data[1] != `text` && data[1] != `int64` && data[1] != `time` && data[1] != `hash` && data[1] != `double` && data[1] != `money` {
			return p.ErrInfo(`incorrect type`)
		}
		if data[2] == `1` {
			if data[1] == `text` {
				return p.ErrInfo(`incorrect index type`)
			}
			indexes++
		}
	}
	if indexes > consts.MAX_INDEXES {
		return fmt.Errorf(`Too many indexes. Limit is %d`, consts.MAX_INDEXES)
	}

	prefix := p.TxStateIDStr
	table := p.TxStateIDStr + `_tables`
	if p.TxMaps.Int64["global"] == 1 {
		table = `global_tables`
		prefix = `global`
	}

	exists, err := p.Single(`SELECT count(*) FROM "`+table+`" WHERE name = ?`, prefix+`_`+p.TxMaps.String["table_name"]).Int64()
	log.Debug(`SELECT count(*) FROM "` + table + `" WHERE name = ?`)
	if err != nil {
		return p.ErrInfo(err)
	}
	if exists > 0 {
		return p.ErrInfo(`table exists`)
	}

	// must be supplemented
	forSign := fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxCitizenID, p.TxStateID, p.TxMap["global"], p.TxMap["table_name"], p.TxMap["columns"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if err := p.AccessRights("new_table", false); err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

// NewTable proceeds NewTable transaction
func (p *Parser) NewTable() error {

	tableName := `global_` + p.TxMaps.String["table_name"]
	if p.TxMaps.Int64["global"] == 0 {
		tableName = p.TxStateIDStr + `_` + p.TxMaps.String["table_name"]
	}
	var cols [][]string
	json.Unmarshal([]byte(p.TxMaps.String["columns"]), &cols)

	//citizenIdStr := utils.Int64ToStr(p.TxCitizenID)
	colsSQL := ""
	colsSQL2 := ""
	sqlIndex := ""
	for _, data := range cols {
		colType := ``
		colDef := ``
		switch data[1] {
		case "text":
			colType = `varchar(102400)`
		case "int64":
			colType = `bigint`
			colDef = `NOT NULL DEFAULT '0'`
		case "time":
			colType = `timestamp`
		case "hash":
			colType = `bytea`
		case "double":
			colType = `double precision`
		case "money":
			colType = `decimal (30, 0)`
			colDef = `NOT NULL DEFAULT '0'`
		}
		colsSQL += `"` + data[0] + `" ` + colType + " " + colDef + " ,\n"
		colsSQL2 += `"` + data[0] + `": "ContractConditions(\"MainCondition\")",`
		if data[2] == "1" {
			sqlIndex += `CREATE INDEX "` + tableName + `_` + data[0] + `_index" ON "` + tableName + `" (` + data[0] + `);`
		}
	}
	colsSQL2 = colsSQL2[:len(colsSQL2)-1]

	sql := `CREATE SEQUENCE "` + tableName + `_id_seq" START WITH 1;
				CREATE TABLE "` + tableName + `" (
				"id" bigint NOT NULL  default nextval('` + tableName + `_id_seq'),
				` + colsSQL + `
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER SEQUENCE "` + tableName + `_id_seq" owned by "` + tableName + `".id;
				ALTER TABLE ONLY "` + tableName + `" ADD CONSTRAINT "` + tableName + `_pkey" PRIMARY KEY (id);`
	fmt.Println(sql)
	err := p.ExecSQL(sql)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSQL(sqlIndex)
	if err != nil {
		return p.ErrInfo(err)
	}

	prefix := `global`
	if p.TxMaps.Int64["global"] == 0 {
		prefix = p.TxStateIDStr
	}
	err = p.ExecSQL(`INSERT INTO "`+prefix+`_tables" ( name, columns_and_permissions ) VALUES ( ?, ? )`,
		tableName, `{"general_update":"ContractConditions(\"MainCondition\")", "update": {`+colsSQL2+`},
		"insert": "ContractConditions(\"MainCondition\")", "new_column":"ContractConditions(\"MainCondition\")"}`)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

// NewTableRollback rollbacks NewTable transaction
func (p *Parser) NewTableRollback() error {

	err := p.autoRollback()
	if err != nil {
		return p.ErrInfo(err)
	}

	tableName := `global_` + p.TxMaps.String["table_name"]
	if p.TxMaps.Int64["global"] == 0 {
		tableName = p.TxStateIDStr + `_` + p.TxMaps.String["table_name"]
	}
	err = p.ExecSQL(`DROP TABLE "` + tableName + `"`)

	prefix := `global`
	if p.TxMaps.Int64["global"] == 0 {
		prefix = p.TxStateIDStr
	}
	err = p.ExecSQL(`DELETE FROM "`+prefix+`_tables" WHERE name = ?`, tableName)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}
