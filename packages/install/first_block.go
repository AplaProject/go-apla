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
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/parser"
	"github.com/GenesisKernel/go-genesis/packages/utils"
)

const fileMode = 0644

func createKeyPair(privFilename, pubFilename string) (priv, pub []byte, err error) {
	priv, pub, err = crypto.GenBytesKeys()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("generate keys")
		return
	}

	err = createFile(privFilename, []byte(hex.EncodeToString(priv)))
	if err != nil {
		return
	}

	err = createFile(pubFilename, []byte(hex.EncodeToString(pub)))
	if err != nil {
		return
	}

	return
}

func createFile(filename string, data []byte) error {
	err := ioutil.WriteFile(filename, data, fileMode)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing file")
		return err
	}
	return nil
}

func generateFirstBlock(publicKey, nodePublicKey []byte) error {
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

	return createFile(*conf.FirstBlockPath, block)
}

// GenerateFirstBlock generates the first block
func GenerateFirstBlock() error {
	var publicKey, nodePublicKey []byte
	var err error

	// publicKey
	if len(*conf.FirstBlockPublicKey) > 0 {
		publicKey, err = hex.DecodeString(*conf.FirstBlockPublicKey)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding key from hex to string")
			return err
		}
	} else {
		_, publicKey, err = createKeyPair(
			filepath.Join(conf.Config.PrivateDir, consts.PrivateKeyFilename),
			filepath.Join(conf.Config.PrivateDir, consts.PublicKeyFilename),
		)
		if err != nil {
			return err
		}
	}

	// nodePublicKey
	if len(*conf.FirstBlockNodePublicKey) > 0 {
		nodePublicKey, err = hex.DecodeString(*conf.FirstBlockNodePublicKey)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding key from hex to string")
			return err
		}
	} else {
		_, nodePublicKey, err = createKeyPair(
			filepath.Join(conf.Config.PrivateDir, consts.NodePrivateKeyFilename),
			filepath.Join(conf.Config.PrivateDir, consts.NodePublicKeyFilename),
		)
		if err != nil {
			return err
		}
	}

	address := crypto.Address(publicKey)
	conf.Config.KeyID = address

	err = createFile(
		filepath.Join(conf.Config.PrivateDir, consts.KeyIDFilename),
		[]byte(strconv.FormatInt(address, 10)),
	)
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
