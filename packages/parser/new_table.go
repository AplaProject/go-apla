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
	"github.com/DayLightProject/go-daylight/packages/script"
	"encoding/json"
)

/*
Adding state tables should be spelled out in state settings
*/

func (p *Parser) NewTableInit() error {

	fields := []map[string]string{{"public_key": "bytes"}, {"table_name": "string"}, {"table_columns": "string"}}
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

	newStateCondition := "#dlt_wallets[wallet_id=walletId].amount > 0"

	vars := map[string]interface{}{
		`citizenId`: 	p.TxCitizenID,
		`walletId`: 	p.TxWalletID,
		`Table`:     	p.MyTable,
	}
	out, err := script.EvalIf(newStateCondition, &vars)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !out {
		return p.ErrInfo("newStateCond false")
	}

	// must be supplemented
	forSign := fmt.Sprintf("%s,%s,%d", p.TxMap["type"], p.TxMap["time"], p.TxMap["state_id"], p.TxCitizenID)
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

	var cols []string
	json.Unmarshal(p.TxMaps.Bytes["table_columns"], &cols)

	colsSql := ""
	for _,name := range cols {
		colsSql += name+" varchar(255) NOT NULL DEFAULT ''\n"
	}

	sql := `CREATE SEQUENCE `+p.TxMaps.String["table_name"]+`_id_seq START WITH 1;
				CREATE TABLE "`+p.TxMaps.String["table_name"]+`" (
				"id" bigint NOT NULL  default nextval('`+p.TxMaps.String["table_name"]+`_id_seq'),
				`+colsSql+`
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER SEQUENCE `+p.TxMaps.String["table_name"]+`_id_seq owned by `+p.TxMaps.String["table_name"]+`.id;
				ALTER TABLE ONLY "`+p.TxMaps.String["table_name"]+`" ADD CONSTRAINT `+p.TxMaps.String["table_name"]+`_pkey PRIMARY KEY (id);`

	err := p.ExecSql(sql)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`INSERT INTO `+p.TxVars[`state_code`]+`_tables ( name, columns_and_permissions ) VALUES ( ?, ? )`,
		p.TxMaps.String["table_name"], p.TxMaps.String["table_columns"])
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) NewTableRollback() error {

	err := p.ExecSql(`DROP TABLE "`+p.TxMaps.String["table_name"]+`"`)

	err = p.ExecSql(`DELETE FROM `+p.TxVars[`state_code`]+
	`_state_tables WHERE name = ?`, p.TxMaps.String["table_name"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewTableRollbackFront() error {

	return nil
}