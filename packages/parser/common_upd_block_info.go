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

	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/AplaProject/go-apla/packages/converter"
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
	forSha := fmt.Sprintf("%d,%x,%s,%d,%d,%d,%d", blockID, block.PrevHeader.Hash, block.MrklRoot,
		block.Header.Time, block.Header.EcosystemID, block.Header.KeyID, block.Header.NodePosition)
	log.Debug("forSha %v", forSha)

	hash, err := crypto.DoubleHash([]byte(forSha))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("block.Header.NodePosition",block.Header.NodePosition)

	block.Header.Hash = hash
	if block.Header.BlockID == 1 {
		ib := &model.InfoBlock{
			Hash:           hash,
			BlockID:        blockID,
			Time:           block.Header.Time,
			EcosystemID:       block.Header.EcosystemID,
			KeyID:       block.Header.KeyID,
			NodePosition:        converter.Int64ToStr(block.Header.NodePosition),
			CurrentVersion: fmt.Sprintf("%d", block.Header.Version),
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
			EcosystemID: block.Header.EcosystemID,
			KeyID: block.Header.KeyID,
			NodePosition:  converter.Int64ToStr(block.Header.NodePosition),
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
