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
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"strings"
)

// откат не всех полей, а только указанных, либо 1 строку, если нет where
func (p *Parser) selectiveRollback(table string, where string, rollbackAI bool) error {
	if len(where) > 0 {
		where = " WHERE " + where
	}
	tblname := lib.EscapeName(table)
	// получим rb_id, по которому можно найти данные, которые были до этого
	rbId, err := p.Single("SELECT rb_id FROM " + tblname + " " + where).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if rbId > 0 {
		// данные, которые восстановим
		rbData, err := p.OneRow("SELECT * FROM rollback WHERE rb_id  =  ?", rbId).String()
		if err != nil {
			return p.ErrInfo(err)
		}

		var jsonMap map[string]string
		json.Unmarshal([]byte(rbData["data"]), &jsonMap)

		//log.Debug("logData",logData)
		addSqlUpdate := ""
		for k, v := range jsonMap {
			if utils.InSliceString(k, []string{"hash", "tx_hash", "public_key_0", "public_key_1", "public_key_2", "node_public_key"}) && len(v) != 0 {
				addSqlUpdate += k + `=decode('` + string(utils.BinToHex([]byte(v))) + `','HEX'),`
			} else {
				addSqlUpdate += k + `='` + strings.Replace(v, `'`, `''`, -1) + `',`
			}
		}
		//log.Debug("%v", logData)
		//log.Debug("%v", logData["prev_rb_id"])
		//log.Debug("UPDATE "+table+" SET "+addSqlUpdate+" rb_id = ? "+where)
		addSqlUpdate = addSqlUpdate[0 : len(addSqlUpdate)-1]
		err = p.ExecSql("UPDATE " + tblname + " SET " + addSqlUpdate + " " + where)
		if err != nil {
			return p.ErrInfo(err)
		}
		// подчищаем _log
		err = p.ExecSql("DELETE FROM rollback WHERE rb_id = ?", rbId)
		if err != nil {
			return p.ErrInfo(err)
		}
		p.rollbackAI("rollback", 1)
	} else {
		err = p.ExecSql("DELETE FROM " + tblname + " " + where)
		if err != nil {
			return p.ErrInfo(err)
		}
		if rollbackAI {
			p.rollbackAI(table, 1)
		}
	}

	return nil
}
