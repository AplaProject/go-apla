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
	"github.com/DayLightProject/go-daylight/packages/utils"
	"fmt"
	"encoding/json"
)

/*
Adding state tables should be spelled out in state settings
*/

func (p *Parser) NewTableInit() error {

	fields := []map[string]string{{"global": "int64"},{"table_name": "string"}, {"columns": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}



func (p *Parser) NewTableFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// Check the system limits. You can not send more than X time a day this TX
	// ...



	// Check InputData
	verifyData := map[string]string{}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// New state table can only add a citizen of the same country
	// ...


	// Check the condition that must be met to complete this transaction
	// select value from ea_state_parameters where name = "new_state_table"
	// ...

	var cols []string
	err = json.Unmarshal([]byte(p.TxMaps.String["columns"]), &cols)
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(cols) == 0 {
		return p.ErrInfo(`len(cols) == 0`)
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

	return nil
}

func (p *Parser) NewTable() error {

	tableName := `global_`+p.TxMaps.String["table_name"]
	if p.TxMaps.Int64["global"] == 0 {
		tableName = p.TxStateIDStr+`_`+p.TxMaps.String["table_name"]
	}
	var cols []string
	json.Unmarshal([]byte(p.TxMaps.String["columns"]), &cols)

	citizenIdStr := utils.Int64ToStr(p.TxCitizenID)
	colsSql := ""
	colsSql2 := ""
	for _,name := range cols {
		colsSql += `"`+name+"\" varchar NOT NULL DEFAULT '',\n"
		colsSql2 += `"`+name+`": "`+p.TxStateIDStr+`_citizens.id=`+citizenIdStr+`",`
	}
	colsSql2 = colsSql2[:len(colsSql2)-1]

	sql := `CREATE SEQUENCE "`+tableName+`_id_seq" START WITH 1;
				CREATE TABLE "`+tableName+`" (
				"id" bigint NOT NULL  default nextval('`+tableName+`_id_seq'),
				`+colsSql+`
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER SEQUENCE "`+tableName+`_id_seq" owned by "`+tableName+`".id;
				ALTER TABLE ONLY "`+tableName+`" ADD CONSTRAINT "`+tableName+`_pkey" PRIMARY KEY (id);`
	fmt.Println(sql)
	err := p.ExecSql(sql)
	if err != nil {
		return p.ErrInfo(err)
	}

	prefix := `global`
	if p.TxMaps.Int64["global"] == 0 {
		prefix = p.TxStateIDStr
	}
	err = p.ExecSql(`INSERT INTO `+prefix+`_tables ( name, columns_and_permissions ) VALUES ( ?, ? )`,
		tableName, `{"general_update":"`+p.TxStateIDStr+`_citizens.id=`+citizenIdStr+`", "update": {`+colsSql2+`}, "insert": "`+p.TxStateIDStr+`_citizens.id=`+citizenIdStr+`", "new_column":"`+p.TxStateIDStr+`_citizens.id=`+citizenIdStr+`"}`)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) NewTableRollback() error {

	tableName := `global_`+p.TxMaps.String["table_name"]
	if p.TxMaps.Int64["global"] == 0 {
		tableName = p.TxStateIDStr+`_`+p.TxMaps.String["table_name"]
	}
	err := p.ExecSql(`DROP TABLE "`+tableName+`"`)

	prefix := `global`
	if p.TxMaps.Int64["global"] == 0 {
		prefix = p.TxStateIDStr
	}
	err = p.ExecSql(`DELETE FROM `+prefix+`_tables WHERE name = ?`, tableName)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewTableRollbackFront() error {
	return nil
}