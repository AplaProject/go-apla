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
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
)

// selectiveRollback rollbacks the specified fields
// откат не всех полей, а только указанных, либо 1 строку, если нет where
// roll back not all the fields but the specified ones or only 1 line if there is not 'where'
func (p *Parser) selectiveRollback(table string, where string, rollbackAI bool) error {
	if len(where) > 0 {
		where = " WHERE " + where
	}
	tblname := converter.EscapeName(table)
	// получим rb_id, по которому можно найти данные, которые были до этого
	// we obtain rb_id with help of that it is possible to find the data which was before
	rbID, err := model.GetRollbackID(tblname, where, "desc")
	if err != nil {
		return p.ErrInfo(err)
	}
	if rbID > 0 {
		// данные, которые восстановим
		// data that we will be restored
		rollback := &model.Rollback{}
		err = rollback.Get(rbID)
		if err != nil {
			return p.ErrInfo(err)
		}

		var jsonMap map[string]string
		err = json.Unmarshal([]byte(rollback.Data), &jsonMap)
		if err != nil {
			return p.ErrInfo(err)
		}
		//log.Debug("logData",logData)
		addSQLUpdate := ""
		for k, v := range jsonMap {
			if converter.InSliceString(k, []string{"hash", "tx_hash", "public_key_0", "node_public_key"}) && len(v) != 0 {
				addSQLUpdate += k + `=decode('` + string(converter.BinToHex([]byte(v))) + `','HEX'),`
			} else {
				addSQLUpdate += k + `='` + strings.Replace(v, `'`, `''`, -1) + `',`
			}
		}
		addSQLUpdate = addSQLUpdate[0 : len(addSQLUpdate)-1]
		err = model.Update(tblname, addSQLUpdate, where)
		if err != nil {
			return p.ErrInfo(err)
		}
		// подчищаем _log
		// clean up the _log
		rbToDel := &model.Rollback{RbID: rbID}
		err = rbToDel.Delete()
		if err != nil {
			return p.ErrInfo(err)
		}
		p.rollbackAI("rollback", 1)
	} else {
		err = model.Delete(tblname, where)
		if err != nil {
			return p.ErrInfo(err)
		}
		if rollbackAI {
			p.rollbackAI(table, 1)
		}
	}

	return nil
}
