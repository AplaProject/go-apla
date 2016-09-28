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

func (p *Parser) NewStateInit() error {

	fields := []map[string]string{{"state_name": "string"}, {"currency_name": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}


func (p *Parser) NewStateFront() error {
	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// Check InputData
	verifyData := map[string]string{"state_name": "string", "currency_name": "string"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	systemParametersJson, err := p.Single(`SELECT value FROM system_parameters WHERE name = ?`, "op_price").String()
	if err != nil {
		return p.ErrInfo(err)
	}
	systemParametersMap := make(map[string]string)
	log.Debug("%s", systemParametersJson)
	err = json.Unmarshal([]byte(systemParametersJson), &systemParametersMap)
	if err != nil {
		return p.ErrInfo(err)
	}
	fPrice := utils.StrToInt64(systemParametersMap["add_state"])

	fuelRate, err := p.Single(`SELECT value FROM system_parameters WHERE name = ?`, "fuel_rate").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	dltPrice := int64(fPrice / fuelRate)

	// есть ли нужная сумма на кошельке
	_, err = p.checkSenderDLT(0, dltPrice)
	if err != nil {
		return p.ErrInfo(err)
	}

	forSign := fmt.Sprintf("%s,%s,%d,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxWalletID, p.TxMap["state_name"], p.TxMap["currency_name"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) NewState() error {

	id_, err := p.ExecSqlGetLastInsertId(`INSERT INTO system_states ( name ) VALUES ( ? )`, "system_states",	p.TxMaps.String["state_name"])
	if err != nil {
		return p.ErrInfo(err)
	}
	id := utils.Int64ToStr(id_)
	err = p.ExecSql("INSERT INTO rollback_tx ( block_id, tx_hash, table_name, table_id ) VALUES (?, [hex], ?, ?)", p.BlockData.BlockId, p.TxHash, "system_states", id)
	if err != nil {
		return err
	}

	err = p.ExecSql(`CREATE TABLE "`+id+`_state_parameters" (
				"name" bytea  NOT NULL DEFAULT '',
				"value" bytea  NOT NULL DEFAULT '',
				"bytecode" bytea  NOT NULL DEFAULT '',
				"conditions" bytea  NOT NULL DEFAULT '',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER TABLE ONLY "`+id+`_state_parameters" ADD CONSTRAINT `+id+`_state_parameters_pkey PRIMARY KEY (name);
				`)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`INSERT INTO "`+id+`_state_parameters" (name, value, bytecode, conditions) VALUES
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?),
		(?, ?, ?, ?)`,
		"main_conditions", id+`_citizens.id=1`, "", "",
		"new_table", id+`_citizens.id=1`, "", id+`_state_parameters.main_conditions`,
		"new_column", id+`_citizens.id=1`, "", id+`_state_parameters.main_conditions`,
		"changing_tables", id+`_citizens.id=1`, "", id+`_state_parameters.main_conditions`,
		"changing_smart_contracts", id+`_citizens.id=1`, "", id+`_state_parameters.main_conditions`,
		"currency_name", p.TxMap["currency_name"], "", id+`_state_parameters.main_conditions`,
		"state_name", p.TxMap["state_name"], "", id+`_state_parameters.main_conditions`,
		"dlt_spending", p.TxWalletID, "", id+`_state_parameters.main_conditions`,
		"citizenship_price", "1000000", "", id+`_state_parameters.main_conditions`)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`CREATE SEQUENCE `+id+`_smart_contracts_id_seq START WITH 1;
				CREATE TABLE "`+id+`_smart_contracts" (
				"id" bigint NOT NULL  default nextval('`+id+`_smart_contracts_id_seq'),
				"name" bytea  NOT NULL DEFAULT '',
				"value" bytea  NOT NULL DEFAULT '',
				"conditions" bytea  NOT NULL DEFAULT '',
				"variables" bytea  NOT NULL DEFAULT '',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER SEQUENCE `+id+`_smart_contracts_id_seq owned by `+id+`_smart_contracts.id;
				ALTER TABLE ONLY "`+id+`_smart_contracts" ADD CONSTRAINT `+id+`_smart_contracts_pkey PRIMARY KEY (id);
				`)
	if err != nil {
		return p.ErrInfo(err)
	}


	err = p.ExecSql(`CREATE TABLE "`+id+`_tables" (
				"name" bytea  NOT NULL DEFAULT '',
				"columns_and_permissions" jsonb,
				"conditions" bytea  NOT NULL DEFAULT '',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER TABLE ONLY "`+id+`_tables" ADD CONSTRAINT `+id+`_tables_pkey PRIMARY KEY (name);
				`)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`INSERT INTO "`+id+`_tables" (name, columns_and_permissions, conditions) VALUES
		(?, ?, ?),
		(?, ?, ?)`,
		id+`_citizens`, `{"general_update":"`+id+`_citizens.id=1", "update": {"public_key": "`+id+`_citizens.id=1"}, "update": {"public_key":""}, "insert": "`+id+`_citizens.id=1", "add_column":"`+id+`_citizens.id=1"}`, id+`_state_parameters.main_conditions`,
		id+`_accounts`, `{"general_update":"`+id+`_citizens.id=1", "update": {"amount": "`+id+`_citizens.id=1"}, "update": {"amount":""}, "insert": "`+id+`_citizens.id=1", "add_column":"`+id+`_citizens.id=1"}`, id+`_state_parameters.main_conditions`)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`CREATE SEQUENCE `+id+`_citizens_id_seq START WITH 1;
				CREATE TABLE "`+id+`_citizens" (
				"id" bigint NOT NULL  default nextval('`+id+`_citizens_id_seq'),
				"public_key" bytea  NOT NULL DEFAULT '',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER SEQUENCE `+id+`_citizens_id_seq owned by `+id+`_citizens.id;
				ALTER TABLE ONLY "`+id+`_citizens" ADD CONSTRAINT `+id+`_citizens_pkey PRIMARY KEY (id);
				`)
	if err != nil {
		return p.ErrInfo(err)
	}

	pKey, err := p.Single(`SELECT public_key_0 FROM dlt_wallets WHERE wallet_id = ?`, p.TxWalletID).Bytes()
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`INSERT INTO "`+id+`_citizens" (public_key) VALUES ([hex])`, utils.BinToHex(pKey))
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`CREATE SEQUENCE `+id+`_accounts_id_seq START WITH 1;
				CREATE TABLE "`+id+`_accounts" (
				"id" bigint NOT NULL  default nextval('`+id+`_accounts_id_seq'),
				"amount" bigint  NOT NULL DEFAULT '0',
				"rb_id" bigint NOT NULL DEFAULT '0'
				);
				ALTER SEQUENCE `+id+`_accounts_id_seq owned by `+id+`_accounts.id;
				ALTER TABLE ONLY "`+id+`_accounts" ADD CONSTRAINT `+id+`_accounts_pkey PRIMARY KEY (id);
				`)
	if err != nil {
		return p.ErrInfo(err)
	}


	return nil
}

func (p *Parser) NewStateRollback() error {

	id_, err := p.Single(`SELECT table_id FROM rollback_tx WHERE tx_hash = [hex] AND table_name = ?`, p.TxHash, "system_states").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	id := utils.Int64ToStr(id_)

	err = p.ExecSql(`DROP TABLE "`+id+`_accounts"`)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`DROP TABLE "`+id+`_citizens"`)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`DROP TABLE "`+id+`_tables"`)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`DROP TABLE "`+id+`_smart_contracts"`)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`DROP TABLE "`+id+`_state_parameters"`)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`DELETE FROM "system_states" WHERE id = ?`, id)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewStateRollbackFront() error {

	return nil
}