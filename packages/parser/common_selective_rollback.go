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

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

// selectiveRollback rollbacks the specified fields
// roll back not all the fields but the specified ones or only 1 line if there is not 'where'
func (p *Parser) selectiveRollback(table string, where string) error {
	logger := p.GetLogger()
	if len(where) > 0 {
		where = " WHERE " + where
	}
	// we obtain rb_id with help of that it is possible to find the data which was before
	rbID, err := model.GetRollbackID(p.DbTransaction, table, where, "desc")
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting rollback id")
		return p.ErrInfo(err)
	}
	if rbID > 0 {
		// data that we will be restored
		rollback := &model.Rollback{}
		_, err = rollback.Get(rbID)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting rollback by id")
			return p.ErrInfo(err)
		}

		var jsonMap map[string]string
		err = json.Unmarshal([]byte(rollback.Data), &jsonMap)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling rollback.Data from json")
			return p.ErrInfo(err)
		}
		addSQLUpdate := ""
		for k, v := range jsonMap {
			if converter.InSliceString(k, []string{"hash", "pub", "tx_hash", "public_key_0", "node_public_key"}) && len(v) != 0 {
				addSQLUpdate += k + `=decode('` + string(converter.BinToHex([]byte(v))) + `','HEX'),`
			} else {
				addSQLUpdate += k + `='` + strings.Replace(v, `'`, `''`, -1) + `',`
			}
		}
		addSQLUpdate = addSQLUpdate[0 : len(addSQLUpdate)-1]
		err = model.Update(p.DbTransaction, table, addSQLUpdate, where)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err, "query": addSQLUpdate}).Error("updating table")
			return p.ErrInfo(err)
		}

		// clean up the _log
		rbToDel := &model.Rollback{RbID: rbID}
		err = rbToDel.Delete()
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting rollback")
			return p.ErrInfo(err)
		}
	} else {
		err = model.Delete(table, where)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting from table")
			return p.ErrInfo(err)
		}
	}

	return nil
}
