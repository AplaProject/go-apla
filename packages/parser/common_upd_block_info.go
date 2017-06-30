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

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// UpdBlockInfo updates info_block table
func (p *Parser) UpdBlockInfo() {

	blockID := p.BlockData.BlockID
	// для локальных тестов
	// for the local tests
	if p.BlockData.BlockID == 1 {
		if *utils.StartBlockID != 0 {
			blockID = *utils.StartBlockID
		}
	}
	forSha := fmt.Sprintf("%d,%s,%s,%d,%d,%d", blockID, p.PrevBlock.Hash, p.MrklRoot, p.BlockData.Time, p.BlockData.WalletID, p.BlockData.StateID)
	log.Debug("forSha", forSha)
	hash, err := crypto.DoubleHash([]byte(forSha))
	if err != nil {
		log.Fatal(err)
	}
	hash = converter.BinToHex(hash)
	p.BlockData.Hash = hash
	log.Debug("%v", p.BlockData.Hash)
	log.Debug("%v", blockID)
	log.Debug("%v", p.BlockData.Time)
	log.Debug("%v", p.CurrentVersion)

	if p.BlockData.BlockID == 1 {
		err := p.ExecSQL("INSERT INTO info_block (hash, block_id, time, state_id, wallet_id, current_version) VALUES ([hex], ?, ?, ?, ?, ?)",
			p.BlockData.Hash, blockID, p.BlockData.Time, p.BlockData.StateID, p.BlockData.WalletID, p.CurrentVersion)
		if err != nil {
			log.Error("%v", err)
		}
	} else {
		err := p.ExecSQL("UPDATE info_block SET hash = [hex], block_id = ?, time = ?, state_id = ?, wallet_id = ?, sent = 0",
			p.BlockData.Hash, blockID, p.BlockData.Time, p.BlockData.StateID, p.BlockData.WalletID)
		if err != nil {
			log.Error("%v", err)
		}
		err = p.ExecSQL("UPDATE config SET my_block_id = ? WHERE my_block_id < ?", blockID, blockID)
		if err != nil {
			log.Error("%v", err)
		}
	}
}
