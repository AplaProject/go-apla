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
	"strconv"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	"gopkg.in/vmihailenco/msgpack.v2"
)

/*
Adding state tables should be spelled out in state settings
*/

type NewTableParser struct {
	*Parser
	NewTable *tx.NewTable
}

func (p *NewTableParser) Init() error {
	newTable := &tx.NewTable{}
	if err := msgpack.Unmarshal(p.TxBinaryData, newTable); err != nil {
		return p.ErrInfo(err)
	}
	p.NewTable = newTable
	return nil
}

func (p *NewTableParser) Validate() error {
	err := p.generalCheck(`add_table`, &p.NewTable.Header)
	if err != nil {
		return p.ErrInfo(err)
	}

	// Check the system limits. You can not send more than X time a day this TX
	// ...

	// Check InputData
	verifyData := map[string][]interface{}{"int64": []interface{}{p.NewTable.Global}, "string": []interface{}{p.NewTable.Name}}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	var cols [][]string
	err = json.Unmarshal([]byte(p.NewTable.Columns), &cols)
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

	prefix := utils.Int64ToStr(p.NewTable.Header.StateID)
	table := prefix + `_tables`
	global, err := strconv.Atoi(p.NewTable.Global)
	if err != nil {
		return fmt.Errorf("Global is not int")
	}
	if global == 1 {
		table = `global_tables`
		prefix = `global`
	}

	exists, err := p.Single(`SELECT count(*) FROM "`+table+`" WHERE name = ?`, prefix+`_`+p.NewTable.Name).Int64()
	log.Debug(`SELECT count(*) FROM "` + table + `" WHERE name = ?`)
	if err != nil {
		return p.ErrInfo(err)
	}
	if exists > 0 {
		return p.ErrInfo(`table exists`)
	}

	// must be supplemented
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.NewTable.ForSign(), p.TxMap["sign"], false)
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

func (p *NewTableParser) Action() error {
	prefix, err := GetTablePrefix(p.NewTable.Global, p.NewTable.Header.StateID)
	if err != nil {
		return p.ErrInfo(err)
	}
	tableName := prefix + "_" + p.NewTable.Name
	var cols [][]string
	json.Unmarshal([]byte(p.NewTable.Columns), &cols)

	colsSql := ""
	colsSql2 := ""
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
		colsSql += `"` + data[0] + `" ` + colType + " " + colDef + " ,\n"
		colsSql2 += `"` + data[0] + `": "ContractConditions(\"MainCondition\")",`
		if data[2] == "1" {
			sqlIndex += `CREATE INDEX "` + tableName + `_` + data[0] + `_index" ON "` + tableName + `" (` + data[0] + `);`
		}
	}
	colsSql2 = colsSql2[:len(colsSql2)-1]

	sql := `CREATE SEQUENCE "` + tableName + `_id_seq" START WITH 1;
				CREATE TABLE "` + tableName + `" (
				"id" bigint NOT NULL  default nextval('` + tableName + `_id_seq'),
				` + colsSql + `
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER SEQUENCE "` + tableName + `_id_seq" owned by "` + tableName + `".id;
				ALTER TABLE ONLY "` + tableName + `" ADD CONSTRAINT "` + tableName + `_pkey" PRIMARY KEY (id);`
	fmt.Println(sql)
	err = p.ExecSql(sql)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(sqlIndex)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`INSERT INTO "`+prefix+`_tables" ( name, columns_and_permissions ) VALUES ( ?, ? )`,
		tableName, `{"general_update":"ContractConditions(\"MainCondition\")", "update": {`+colsSql2+`},
		"insert": "ContractConditions(\"MainCondition\")", "new_column":"ContractConditions(\"MainCondition\")"}`)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *NewTableParser) Rollback() error {
	err := p.autoRollback()
	if err != nil {
		return p.ErrInfo(err)
	}
	prefix, err := GetTablePrefix(p.NewTable.Global, p.NewTable.Header.StateID)
	if err != nil {
		return p.ErrInfo(err)
	}
	tableName := prefix + "_" + p.NewTable.Name
	err = p.ExecSql(`DROP TABLE "` + tableName + `"`)
	err = p.ExecSql(`DELETE FROM "`+prefix+`_tables" WHERE name = ?`, tableName)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p NewTableParser) Header() *tx.Header {
	return &p.NewTable.Header
}
