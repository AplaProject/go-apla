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

	"github.com/DayLightProject/go-daylight/packages/utils"
)

// не использовать для комментов
func (p *Parser) selectiveLoggingAndUpd(fields []string, values_ []interface{}, table string, whereFields, whereValues []string, generalRollback bool) error {

	var tableId int64
	values := utils.InterfaceSliceToStr(values_)

	addSqlFields := p.AllPkeys[table]
	if len(addSqlFields) > 0 {
		addSqlFields += `,`
	}
	for _, field := range fields {
		addSqlFields += field + ","
	}

	addSqlWhere := ""
	if whereFields != nil && whereValues != nil {
		for i := 0; i < len(whereFields); i++ {
			addSqlWhere += whereFields[i] + "=" + whereValues[i] + " AND "
		}
	}
	if len(addSqlWhere) > 0 {
		addSqlWhere = " WHERE " + addSqlWhere[0:len(addSqlWhere)-5]
	}
	// если есть, что логировать
	logData, err := p.OneRow("SELECT " + addSqlFields + " rb_id FROM " + table + " " + addSqlWhere).String()
	if err != nil {
		return err
	}
	if len(logData) > 0 {
		var jsonMap map[string]string
		for k, v := range logData {
			if k == p.AllPkeys[table] {
				continue
			}
			if utils.InSliceString(k, []string{"address", "hash", "tx_hash", "public_key_0", "public_key_1", "public_key_2", "node_public_key"}) && v != "" {
				jsonMap[k] = string(utils.BinToHex([]byte(v)))
			} else {
				jsonMap[k] = v
			}
			if k == "rb_id" {
				k = "prev_rb_id"
			}
			if k[:1] == "+" {
				addSqlFields += k[1:len(k)] + ","
			} else {
				addSqlFields += k + ","
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
		addSqlUpdate := ""
		for i := 0; i < len(fields); i++ {
			if utils.InSliceString(fields[i], []string{"address", "hash", "tx_hash", "public_key_0", "public_key_1", "public_key_2", "node_public_key"}) && len(values[i]) != 0 {
				query := ""
				switch p.ConfigIni["db_type"] {
				case "sqlite":
					query = fields[i] + `=x'` + values[i] + `',`
				case "postgresql":
					query = fields[i] + `=decode('` + values[i] + `','HEX'),`
				case "mysql":
					query = fields[i] + `=UNHEX("` + values[i] + `"),`
				}
				addSqlUpdate += query
			} else if fields[i][:1] == "+" {
				addSqlUpdate += fields[i][1:len(fields[i])] + `='` + fields[i][1:len(fields[i])] + `+` + values[i] + `',`
			} else {
				addSqlUpdate += fields[i] + `='` + values[i] + `',`
			}
		}
		err = p.ExecSql("UPDATE "+table+" SET "+addSqlUpdate+" rb_id = ? "+addSqlWhere, rbId)
		//log.Debug("UPDATE "+table+" SET "+addSqlUpdate+" rb_id = ? "+addSqlWhere)
		//log.Debug("logId", logId)
		if err != nil {
			return err
		}
		tableId = utils.StrToInt64(logData[p.AllPkeys[table]])
	} else {
		addSqlIns0 := ""
		addSqlIns1 := ""
		for i := 0; i < len(fields); i++ {
			addSqlIns0 += `` + fields[i] + `,`
			if utils.InSliceString(fields[i], []string{"hash", "tx_hash", "public_key_0", "public_key_1", "public_key_2", "node_public_key"}) && len(values[i]) != 0 {
				query := ""
				switch p.ConfigIni["db_type"] {
				case "sqlite":
					query = `x'` + values[i] + `',`
				case "postgresql":
					query = `decode('` + values[i] + `','HEX'),`
				case "mysql":
					query = `UNHEX("` + values[i] + `"),`
				}
				addSqlIns1 += query
			} else {
				addSqlIns1 += `'` + values[i] + `',`
			}
		}
		for i := 0; i < len(whereFields); i++ {
			addSqlIns0 += `` + whereFields[i] + `,`
			addSqlIns1 += `'` + whereValues[i] + `',`
		}
		addSqlIns0 = addSqlIns0[0 : len(addSqlIns0)-1]
		addSqlIns1 = addSqlIns1[0 : len(addSqlIns1)-1]
		tableId, err = p.ExecSqlGetLastInsertId("INSERT INTO "+table+" ("+addSqlIns0+") VALUES ("+addSqlIns1+")", table)
		if err != nil {
			return err
		}
	}
	if generalRollback {
		err = p.ExecSql("INSERT INTO rollback_tx ( block_id, tx_hash, table_name, table_id ) VALUES (?, [hex], ?, ?)", p.BlockData.BlockId, p.TxHash, table, tableId)
		if err != nil {
			return err
		}
	}
	return nil
}
