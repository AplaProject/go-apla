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
func UpdBlockInfo(dbTransaction *model.DbTransaction, block *Block) error {
	blockID := block.Header.BlockID
	// for the local tests
	if block.Header.BlockID == 1 {
		if *utils.StartBlockID != 0 {
			blockID = *utils.StartBlockID
		}
	}
	forSha := fmt.Sprintf("%d,%s,%s,%d,%d,%d", blockID, block.PrevHeader.Hash, block.MrklRoot,
		block.Header.Time, block.Header.WalletID, block.Header.StateID)
	log.Debug("forSha", forSha)
	hash, err := crypto.DoubleHash([]byte(forSha))
	if err != nil {
		log.Fatal(err)
	}

	if block.Header.BlockID == 1 {
		ib := &model.InfoBlock{
			Hash:           hash,
			BlockID:        blockID,
			Time:           block.Header.Time,
			StateID:        block.Header.StateID,
			WalletID:       block.Header.WalletID,
			CurrentVersion: block.Version,
		}
		err := ib.Create(dbTransaction)
		if err != nil {
			return fmt.Errorf("error insert into info_block %s", err)
		}
	} else {
		ibUpdate := &model.InfoBlock{
			Hash:     hash,
			BlockID:  blockID,
			Time:     block.Header.Time,
			StateID:  block.Header.StateID,
			WalletID: block.Header.WalletID,
			Sent:     0,
		}
		if err := ibUpdate.Update(dbTransaction); err != nil {
			return fmt.Errorf("error while updating info_block: %s", err)
		}

		config := &model.Config{}
		err = config.ChangeBlockIDBatch(dbTransaction, blockID, blockID)
		if err != nil {
			return err
		}
	}

	return nil
}
