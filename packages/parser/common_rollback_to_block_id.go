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
	"database/sql"
	"fmt"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/logging"
)

// RollbackToBlockID rollbacks blocks till blockID
func (p *Parser) RollbackToBlockID(blockID int64) error {

	/*err := p.ExecSQL("SET GLOBAL net_read_timeout = 86400")
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSQL("SET GLOBAL max_connections  = 86400")
	if err != nil {
		return p.ErrInfo(err)
	}*/
	/*err := p.RollbackTransactions()
	if err != nil {
		return p.ErrInfo(err)
	}*/
	err := p.ExecSQL("UPDATE transactions SET verified = 0 WHERE verified = 1 AND used = 0")
	if err != nil {
		logging.WriteSelectiveLog(err)
		return p.ErrInfo(err)
	}

	limit := 1000
	blocks := make([]map[string][]byte, 0, limit)
	//	var blocks []map[string][]byte
	// откатываем наши блоки
	// roll back our blocks
	for {
		rows, err := p.Query(p.FormatQuery("SELECT id, data FROM block_chain WHERE id > ? ORDER BY id DESC LIMIT "+fmt.Sprintf(`%d`, limit)+` OFFSET 0`), blockID)
		if err != nil {
			return p.ErrInfo(err)
		}
		parser := new(Parser)
		parser.DCDB = p.DCDB
		for rows.Next() {
			var data, id []byte
			err = rows.Scan(&id, &data)
			if err != nil {
				rows.Close()
				return p.ErrInfo(err)
			}
			blocks = append(blocks, map[string][]byte{"id": id, "data": data})
		}
		rows.Close()
		if len(blocks) == 0 {
			break
		}
		fmt.Printf(`%s `, blocks[0]["id"])
		for _, block := range blocks {
			// Откатываем наши блоки до блока blockID
			// roll back our blocks to the block blockID
			parser.BinaryData = block["data"]
			err = parser.ParseDataRollback()
			if err != nil {
				return p.ErrInfo(err)
			}

			err = p.ExecSQL("DELETE FROM block_chain WHERE id = ?", block["id"])
			if err != nil {
				return p.ErrInfo(err)
			}
		}
		blocks = blocks[:0]
	}
	var hash, data []byte
	err = p.QueryRow(p.FormatQuery("SELECT hash, data FROM block_chain WHERE id  =  ?"), blockID).Scan(&hash, &data)
	if err != nil && err != sql.ErrNoRows {
		return p.ErrInfo(err)
	}
	converter.BytesShift(&data, 1)
	iblock := converter.BinToDecBytesShift(&data, 4)
	time := converter.BinToDecBytesShift(&data, 4)
	size, err := converter.DecodeLength(&data)
	if err != nil {
		log.Fatal(err)
	}
	walletID := converter.BinToDecBytesShift(&data, size)
	StateID := converter.BinToDecBytesShift(&data, 1)
	err = p.ExecSQL("UPDATE info_block SET hash = [hex], block_id = ?, time = ?, wallet_id = ?, state_id = ?",
		converter.BinToHex(hash), iblock, time, walletID, StateID)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSQL("UPDATE config SET my_block_id = ?", iblock)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}
