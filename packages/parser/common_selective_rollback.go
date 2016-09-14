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
)

// откат не всех полей, а только указанных, либо 1 строку, если нет where
func (p *Parser) selectiveRollback(fields []string, table string, where string, rollback bool) error {
	if len(where) > 0 {
		where = " WHERE " + where
	}
	addSqlFields := ""
	for _, field := range fields {
		addSqlFields += field + ","
	}
	// получим rb_id, по которому можно найти данные, которые были до этого
	logId, err := p.Single("SELECT rb_id FROM " + table + " " + where).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if logId > 0 {
		// данные, которые восстановим
		logData, err := p.OneRow("SELECT "+addSqlFields+" prev_rb_id FROM rb_"+table+" WHERE rb_id  =  ?", logId).String()
		if err != nil {
			return p.ErrInfo(err)
		}
		//log.Debug("logData",logData)
		addSqlUpdate := ""
		for _, field := range fields {
			if utils.InSliceString(field, []string{"hash", "tx_hash", "public_key_0", "public_key_1", "public_key_2", "node_public_key"}) && len(logData[field]) != 0 {
				query := ""
				logData[field] = string(utils.BinToHex([]byte(logData[field])))
				switch p.ConfigIni["db_type"] {
				case "sqlite":
					query = field + `=x'` + logData[field] + `',`
				case "postgresql":
					query = field + `=decode('` + logData[field] + `','HEX'),`
				case "mysql":
					query = field + `=UNHEX("` + logData[field] + `"),`
				}
				addSqlUpdate += query
			} else {
				addSqlUpdate += field + `='` + logData[field] + `',`
			}
		}
		//log.Debug("%v", logData)
		//log.Debug("%v", logData["prev_rb_id"])
		//log.Debug("UPDATE "+table+" SET "+addSqlUpdate+" rb_id = ? "+where)
		err = p.ExecSql("UPDATE "+table+" SET "+addSqlUpdate+" rb_id = ? "+where, logData["prev_rb_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		// подчищаем _log
		err = p.ExecSql("DELETE FROM rb_"+table+" WHERE rb_id = ?", logId)
		if err != nil {
			return p.ErrInfo(err)
		}
		p.rollbackAI("rb_"+table, 1)
	} else {
		err = p.ExecSql("DELETE FROM " + table + " " + where)
		if err != nil {
			return p.ErrInfo(err)
		}
		if rollback {
			p.rollbackAI(table, 1)
		}
	}

	return nil
}