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
	"fmt"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

func (p *Parser) generalRollback(table string, whereUserID interface{}, addWhere string, AI bool) error {
	var UserID int64
	switch whereUserID.(type) {
	case string:
		UserID = utils.StrToInt64(whereUserID.(string))
	case []byte:
		UserID = utils.BytesToInt64(whereUserID.([]byte))
	case int:
		UserID = int64(whereUserID.(int))
	case int64:
		UserID = whereUserID.(int64)
	}

	where := ""
	if UserID > 0 {
		where = fmt.Sprintf(" WHERE user_id = %d ", UserID)
	}
	// получим rb_id, по которому можно найти данные, которые были до этого
	// will get rb_id with help of which it is possible to find data that was before
	logID, err := p.Single("SELECT rb_id FROM " + table + " " + where + addWhere).Int64()
	if err != nil {
		return utils.ErrInfo(err)
	}
	// если $rb_id = 0, значит восстанавливать нечего и нужно просто удалить запись
	// if $rb_id = 0, then there is nothing to restore and you just need to delete the record
	if logID == 0 {
		err = p.ExecSql("DELETE FROM " + table + " " + where + addWhere)
		if err != nil {
			return utils.ErrInfo(err)
		}
	} else {
		// данные, которые восстановим
		// data that will be restored
		data, err := p.OneRow("SELECT * FROM rb_"+table+" WHERE rb_id = ?", logID).String()
		if err != nil {
			return utils.ErrInfo(err)
		}
		addSQL := ""
		for k, v := range data {
			// block_id т.к. в rb_ он нужен для удаления старых данных, а в обычной табле не нужен
			// block_id (because it is needed for removement of old data in rb_ but in usual table there is no need in it)
			if k == "rb_id" || k == "prev_rb_id" || k == "block_id" {
				continue
			}
			if k == "node_public_key" {
				switch p.ConfigIni["db_type"] {
				case "sqlite":
					addSQL += fmt.Sprintf("%v='%x',", k, v)
				case "postgresql":
					addSQL += fmt.Sprintf("%v=decode('%x','HEX'),", k, v)
				case "mysql":
					addSQL += fmt.Sprintf("%v=UNHEX('%x'),", k, v)
				}
			} else {
				addSQL += fmt.Sprintf("%v = '%v',", k, v)
			}
		}
		// всегда пишем предыдущий rb_id
		// always write the previous rb_id
		addSQL += fmt.Sprintf("rb_id = %v,", data["prev_rb_id"])
		addSQL = addSQL[0 : len(addSQL)-1]
		err = p.ExecSql("UPDATE " + table + " SET " + addSQL + where + addWhere)
		if err != nil {
			return utils.ErrInfo(err)
		}
		// подчищаем log
		// Clean up log
		err = p.ExecSql("DELETE FROM rb_"+table+" WHERE rb_id= ?", logID)
		if err != nil {
			return utils.ErrInfo(err)
		}
		err = p.rollbackAI("rb_"+table, 1)
		if err != nil {
			return utils.ErrInfo(err)
		}
	}
	return nil
}
