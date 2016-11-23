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
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

type rollbackTxRowType struct {
	tx_hash    string
	table_name string
	table_id   string
}

func (p *Parser) autoRollback() error {

	var rollbackTxRow rollbackTxRowType
	rows, err := p.QueryRows("SELECT tx_hash, table_name, table_id FROM rollback_tx WHERE tx_hash = [hex] ORDER BY id DESC", p.TxHash)
	if err != nil {
		return utils.ErrInfo(err)
	}
	for rows.Next() {
		err = rows.Scan(&rollbackTxRow.tx_hash, &rollbackTxRow.table_name, &rollbackTxRow.table_id)
		if err != nil {
			return utils.ErrInfo(err)
		}
		err := p.selectiveRollback(rollbackTxRow.table_name, p.AllPkeys[rollbackTxRow.table_name]+"='"+rollbackTxRow.table_id+`'`, true)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	err = p.ExecSql("DELETE FROM rollback_tx WHERE tx_hash = [hex]", p.TxHash)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}
