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

	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// UpdBlockInfo updates info_block table
func (p *Parser) UpdBlockInfo() {
	blockID := p.BlockData.BlockID

	// for the local tests
	if p.BlockData.BlockID == 1 {
		if *utils.StartBlockID != 0 {
			blockID = *utils.StartBlockID
		}
	}
	forSha := fmt.Sprintf("%d,%s,%s,%d,%d,%d", blockID, p.PrevBlock.Hash, p.MrklRoot, p.BlockData.Time, p.BlockData.WalletID, p.BlockData.StateID)
	hash, err := crypto.DoubleHash([]byte(forSha))
	if err != nil {
		log.Fatal(err)
	}

	p.BlockData.Hash = hash

	if p.BlockData.BlockID == 1 {
		ib := &model.InfoBlock{
			Hash:           p.BlockData.Hash,
			BlockID:        blockID,
			Time:           p.BlockData.Time,
			StateID:        p.BlockData.StateID,
			WalletID:       p.BlockData.WalletID,
			CurrentVersion: p.CurrentVersion,
		}
		err := ib.Create()
		if err != nil {
			log.Error("error insert into info_block %v", err)
		}
	} else {
		ibUpdate := &model.InfoBlock{
			Hash:     p.BlockData.Hash,
			BlockID:  blockID,
			Time:     p.BlockData.Time,
			StateID:  p.BlockData.StateID,
			WalletID: p.BlockData.WalletID,
			Sent:     0,
		}
		if err := ibUpdate.Update(); err != nil {
			log.Error("info block update error: %s", err)
		}
		config := &model.Config{}
		err = config.ChangeBlockIDBatch(blockID, blockID)
		if err != nil {
			log.Error("change block id batch error: %s", err)
		}
	}
}
