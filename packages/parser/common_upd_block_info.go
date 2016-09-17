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
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) UpdBlockInfo() {

	blockId := p.BlockData.BlockId
	// для локальных тестов
	if p.BlockData.BlockId == 1 {
		if *utils.StartBlockId != 0 {
			blockId = *utils.StartBlockId
		}
	}
	forSha := fmt.Sprintf("%d,%s,%s,%d,%d,%d", blockId, p.PrevBlock.Hash, p.MrklRoot, p.BlockData.Time, p.BlockData.WalletId, p.BlockData.CBID)
	log.Debug("forSha", forSha)
	p.BlockData.Hash = utils.DSha256(forSha)
	log.Debug("%v", p.BlockData.Hash)
	log.Debug("%v", blockId)
	log.Debug("%v", p.BlockData.Time)
	log.Debug("%v", p.CurrentVersion)

	if p.BlockData.BlockId == 1 {
		err := p.ExecSql("INSERT INTO info_block (hash, block_id, time, state_id, wallet_id, current_version) VALUES ([hex], ?, ?, ?, ?, ?)",
			p.BlockData.Hash, blockId, p.BlockData.Time, p.BlockData.CBID, p.BlockData.WalletId, p.CurrentVersion)
		if err!=nil {
			log.Error("%v", err)
		}
	} else {
		err := p.ExecSql("UPDATE info_block SET hash = [hex], block_id = ?, time = ?, state_id = ?, wallet_id = ?, sent = 0",
			p.BlockData.Hash, blockId, p.BlockData.Time, p.BlockData.CBID, p.BlockData.WalletId)
		if err!=nil {
			log.Error("%v", err)
		}
		err = p.ExecSql("UPDATE config SET my_block_id = ? WHERE my_block_id < ?", blockId, blockId)
		if err!=nil {
			log.Error("%v", err)
		}
	}
}
