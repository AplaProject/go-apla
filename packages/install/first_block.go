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

package install

import (
	"errors"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/parser"
	"github.com/GenesisKernel/go-genesis/packages/utils"
)

const fileMode = 0644

// ErrFirstBlockHostIsEmpty host for first block is not specified
var ErrFirstBlockHostIsEmpty = errors.New("FirstBlockHost is empty")

func generateFirstBlock(publicKey, nodePublicKey []byte) error {
	if len(*conf.FirstBlockHost) == 0 {
		return ErrFirstBlockHostIsEmpty
	}

	now := time.Now().Unix()

	header := &utils.BlockData{
		BlockID:      1,
		Time:         now,
		EcosystemID:  0,
		KeyID:        conf.Config.KeyID,
		NodePosition: 0,
		Version:      consts.BLOCK_VERSION,
	}

	var tx []byte
	_, err := converter.BinMarshal(&tx,
		&consts.FirstBlock{
			TxHeader: consts.TxHeader{
				// TODO: move types to enum
				Type: 1, // FirstBlock

				Time:  uint32(now),
				KeyID: conf.Config.KeyID,
			},
			PublicKey:     publicKey,
			NodePublicKey: nodePublicKey,
			Host:          *conf.FirstBlockHost,
		},
	)

	if err != nil {
		log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("first block body bin marshalling")
		return err
	}

	block, err := parser.MarshallBlock(header, [][]byte{tx}, []byte("0"), "")
	if err != nil {
		log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("first block marshalling")
		return err
	}

	return utils.CreateFile(*conf.FirstBlockPath, block)
}

// GenerateFirstBlock generates the first block
func GenerateFirstBlock() error {
	publicKey, nodePublicKey, err := utils.GenerateKeyFiles()
	if err != nil {
		return err
	}

	return generateFirstBlock(publicKey, nodePublicKey)
}

// IsExistFirstBlock returns a boolean indicating whether first block file exists
func IsExistFirstBlock() bool {
	if _, err := os.Stat(*conf.FirstBlockPath); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
