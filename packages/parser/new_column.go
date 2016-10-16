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
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) NewColumnInit() error {

	fields := []map[string]string{{"table_name": "string"}, {"column_name": "string"}, {"permissions": "string"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewColumnFront() error {
	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// Check InputData
	/*verifyData := map[string]string{"table_name": "string", "column_name": "word", "permissions": "string"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}*/

	fPrice, err := p.Single(`SELECT value->'new_column' FROM system_parameters WHERE name = ?`, "op_price").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

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

	table := p.TxStateIDStr + `_tables`
	exists, err := p.Single(`select count(*) from "`+table+`" where (columns_and_permissions->'update'-> ? ) is not null AND name = ?`, p.TxMaps.String["column_name"], p.TxMaps.String["table_name"]).Int64()
	log.Debug(`select count(*) from "`+table+`" where (columns_and_permissions->'update'-> ? ) is not null AND name = ?`, p.TxMaps.String["column_name"], p.TxMaps.String["table_name"])
	if err != nil {
		return p.ErrInfo(err)
	}
	if exists > 0 {
		return p.ErrInfo(`column exists`)
	}

	forSign := fmt.Sprintf("%s,%s,%d,%d,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxCitizenID, p.TxStateID, p.TxMap["table_name"], p.TxMap["column_name"], p.TxMap["permissions"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) NewColumn() error {

	table := p.TxStateIDStr + `_tables`
	logData, err := p.OneRow(`SELECT columns_and_permissions, rb_id FROM "` + table + `"`).String()
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
	err = p.ExecSql(`UPDATE "`+table+`" SET columns_and_permissions = jsonb_set(columns_and_permissions, '{update, `+p.TxMaps.String["column_name"]+`}', ?, true), rb_id = ? WHERE name = ?`, `"`+p.TxMaps.String["permissions"]+`"`, rbId, p.TxMaps.String["table_name"])
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql("INSERT INTO rollback_tx ( block_id, tx_hash, table_name, table_id ) VALUES (?, [hex], ?, ?)", p.BlockData.BlockId, p.TxHash, table, p.TxMaps.String["table_name"])
	if err != nil {
		return err
	}

	err = p.ExecSql(`ALTER TABLE "`+p.TxMaps.String["table_name"]+`" ADD COLUMN `+p.TxMaps.String["column_name"]+` varchar(10240)`)
	if err != nil {
		return err
	}


	return nil
}

func (p *Parser) NewColumnRollback() error {
	err := p.autoRollback()
	if err != nil {
		return err
	}
	err = p.ExecSql(`ALTER TABLE "`+p.TxMaps.String["table_name"]+`" DROP COLUMN `+p.TxMaps.String["column_name"]+``)
	if err != nil {
		return err
	}
	return nil
}

/*func (p *Parser) NewColumnRollbackFront() error {

	return nil
}
*/